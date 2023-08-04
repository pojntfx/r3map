package migration

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	bbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

type rpcReaderAt struct {
	ctx context.Context

	remote *services.SeederRemote
}

func (b *rpcReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	r, err := b.remote.ReadAt(b.ctx, len(p), off)
	if err != nil {
		return -1, err
	}

	n = r.N
	copy(p, r.P)

	return
}

var (
	ErrStartingTrackFailed = errors.New("starting track failed")
)

type LeecherOptions struct {
	ChunkSize int64

	PullWorkers  int64
	PullPriority func(off int64) int64

	Verbose bool
}

type LeecherHooks struct {
	OnAfterSync func(dirtyOffsets []int64) error

	OnBeforeClose func() error

	OnChunkIsLocal func(off int64) error
}

type PathLeecher struct {
	ctx context.Context

	local  backend.Backend
	remote *services.SeederRemote

	options *LeecherOptions
	hooks   *LeecherHooks

	serverOptions *server.Options
	clientOptions *client.Options

	serverFile *os.File
	dev        *mount.DirectPathMount

	devicePath           string
	syncedReadWriter     *chunks.SyncedReadWriterAt
	puller               *chunks.Puller
	syncer               backend.Backend
	lockableReadWriterAt *chunks.LockableReadWriterAt

	finalizedCond *sync.Cond
	finalized     bool

	pendingChunks sync.WaitGroup

	wg   sync.WaitGroup
	errs chan error
}

func NewPathLeecher(
	ctx context.Context,

	local backend.Backend,
	remote *services.SeederRemote,

	options *LeecherOptions,
	hooks *LeecherHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *PathLeecher {
	if options == nil {
		options = &LeecherOptions{}
	}

	if options.ChunkSize <= 0 {
		options.ChunkSize = 4096
	}

	if options.PullWorkers <= 0 {
		options.PullWorkers = 512
	}

	if options.PullPriority == nil {
		options.PullPriority = func(off int64) int64 {
			return 1
		}
	}

	if hooks == nil {
		hooks = &LeecherHooks{}
	}

	return &PathLeecher{
		ctx: ctx,

		local:  local,
		remote: remote,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		finalizedCond: sync.NewCond(&sync.Mutex{}),
		finalized:     false,

		errs: make(chan error),
	}
}

func (l *PathLeecher) Wait() error {
	for err := range l.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *PathLeecher) Open() (int64, error) {
	ready := make(chan struct{})

	go func() {
		if err := l.remote.Track(l.ctx); err != nil {
			l.errs <- err

			close(ready)

			return
		}

		ready <- struct{}{}
	}()

	size, err := l.local.Size()
	if err != nil {
		return 0, err
	}

	l.devicePath, err = utils.FindUnusedNBDDevice()
	if err != nil {
		return 0, err
	}

	l.serverFile, err = os.Open(l.devicePath)
	if err != nil {
		return 0, err
	}

	chunkCount := size / l.options.ChunkSize
	l.pendingChunks.Add(int(chunkCount))

	hook := l.hooks.OnChunkIsLocal
	l.syncedReadWriter = chunks.NewSyncedReadWriterAt(&rpcReaderAt{l.ctx, l.remote}, l.local, func(off int64) error {
		l.pendingChunks.Done()

		if hook != nil {
			return hook(off)
		}

		return nil
	})

	l.puller = chunks.NewPuller(
		l.ctx,
		l.syncedReadWriter,
		l.options.ChunkSize,
		chunkCount,
		func(off int64) int64 {
			return l.options.PullPriority(off)
		},
	)

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		if err := l.puller.Wait(); err != nil {
			l.errs <- err

			return
		}
	}()

	_, ok := <-ready
	if !ok {
		return 0, ErrStartingTrackFailed
	}

	if err := l.puller.Open(l.options.PullWorkers); err != nil {
		return 0, err
	}

	l.lockableReadWriterAt = chunks.NewLockableReadWriterAt(
		chunks.NewArbitraryReadWriterAt(l.syncedReadWriter, l.options.ChunkSize),
	)

	l.syncer = bbackend.NewReaderAtBackend(
		l.lockableReadWriterAt,
		func() (int64, error) {
			return size, nil
		},
		func() error {
			// We don't need to call `Sync()` here since our remote handles it when we call `Finalize()`

			return nil
		},
		l.options.Verbose,
	)

	l.dev = mount.NewDirectPathMount(
		l.syncer,
		l.serverFile,

		l.serverOptions,
		l.clientOptions,
	)

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		if err := l.dev.Wait(); err != nil {
			l.errs <- err

			return
		}
	}()

	if err := l.dev.Open(); err != nil {
		return 0, err
	}

	return size, nil
}

func (l *PathLeecher) Finalize() (string, error) {
	dirtyOffsets, err := l.remote.Sync(l.ctx)
	if err != nil {
		return "", err
	}

	l.pendingChunks.Add(len(dirtyOffsets))

	if hook := l.hooks.OnAfterSync; hook != nil {
		if err := hook(dirtyOffsets); err != nil {
			return "", err
		}
	}

	if l.syncedReadWriter != nil {
		l.syncedReadWriter.MarkAsRemote(dirtyOffsets)
	}

	if l.puller != nil {
		l.puller.Finalize(dirtyOffsets)
	}

	l.lockableReadWriterAt.Unlock()

	l.finalizedCond.L.Lock()
	l.finalized = true
	l.finalizedCond.Broadcast()
	l.finalizedCond.L.Unlock()

	return l.devicePath, nil
}

func (l *PathLeecher) Release() error {
	l.finalizedCond.L.Lock()
	if !l.finalized {
		l.finalizedCond.Wait()
	}
	l.finalizedCond.L.Unlock()

	l.pendingChunks.Wait()

	return nil
}

func (l *PathLeecher) Close() error {
	if hook := l.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	if l.dev != nil {
		_ = l.dev.Close()
	}

	if l.puller != nil {
		_ = l.puller.Close()
	}

	if l.serverFile != nil {
		_ = l.serverFile.Close()
	}

	_ = l.remote.Close(l.ctx)

	l.wg.Wait()

	if l.errs != nil {
		close(l.errs)

		l.errs = nil
	}

	return nil
}
