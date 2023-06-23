package migration

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	bbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/utils"
)

var (
	ErrStartingTrackFailed = errors.New("starting track failed")
)

type DestinationOptions struct {
	ChunkSize int64

	PullWorkers int64

	Verbose bool
}

type DestinationHooks struct {
	OnChunkIsLocal func(off int64) error
	OnAfterFlush   func(dirtyOffsets []int64) error
}

type Destination struct {
	ctx context.Context

	remote io.ReaderAt
	size   int64

	track func() error
	flush func() ([]int64, error)
	close func() error

	local,
	syncer backend.Backend

	options *DestinationOptions
	hooks   *DestinationHooks

	serverOptions *server.Options
	clientOptions *client.Options

	serverFile       *os.File
	syncedReadWriter *chunks.SyncedReadWriterAt
	puller           *chunks.Puller
	dev              *device.Device

	wg   sync.WaitGroup
	errs chan error
}

func NewDestination(
	ctx context.Context,

	remote io.ReaderAt,
	size int64,

	track func() error,
	flush func() ([]int64, error),
	close func() error,

	local backend.Backend,

	options *DestinationOptions,
	hooks *DestinationHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *Destination {
	if options == nil {
		options = &DestinationOptions{}
	}

	if options.ChunkSize <= 0 {
		options.ChunkSize = 4096
	}

	if options.PullWorkers <= 0 {
		options.PullWorkers = 512
	}

	if hooks == nil {
		hooks = &DestinationHooks{}
	}

	return &Destination{
		ctx: ctx,

		remote: remote,
		size:   size,

		track: track,
		flush: flush,
		close: close,

		local: local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (m *Destination) Wait() error {
	for err := range m.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Destination) Open() (string, error) {
	ready := make(chan struct{})

	go func() {
		if err := m.track(); err != nil {
			m.errs <- err

			close(ready)

			return
		}

		ready <- struct{}{}
	}()

	chunkCount := m.size / m.options.ChunkSize

	devicePath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		return "", err
	}

	m.serverFile, err = os.Open(devicePath)
	if err != nil {
		return "", err
	}

	local := chunks.NewChunkedReadWriterAt(m.local, m.options.ChunkSize, chunkCount)

	hook := m.hooks.OnChunkIsLocal
	m.syncedReadWriter = chunks.NewSyncedReadWriterAt(m.remote, local, func(off int64) error {
		if hook != nil {
			return hook(off)
		}

		return nil
	})

	m.puller = chunks.NewPuller(
		m.ctx,
		m.syncedReadWriter,
		m.options.ChunkSize,
		chunkCount,
		func(offset int64) int64 {
			return 1
		},
	)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		if err := m.puller.Wait(); err != nil {
			m.errs <- err

			return
		}
	}()

	_, ok := <-ready
	if !ok {
		return "", ErrStartingTrackFailed
	}

	if err := m.puller.Open(m.options.PullWorkers); err != nil {
		return "", err
	}

	arbitraryReadWriter := chunks.NewArbitraryReadWriterAt(m.syncedReadWriter, m.options.ChunkSize)

	m.syncer = bbackend.NewReaderAtBackend(
		arbitraryReadWriter,
		func() (int64, error) {
			return m.size, nil
		},
		func() error {
			if err := m.local.Sync(); err != nil {
				return err
			}

			return nil
		},
		m.options.Verbose,
	)

	m.dev = device.NewDevice(
		m.syncer,
		m.serverFile,

		m.serverOptions,
		m.clientOptions,
	)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		if err := m.dev.Wait(); err != nil {
			m.errs <- err

			return
		}
	}()

	if err := m.dev.Open(); err != nil {
		return "", err
	}

	return devicePath, nil
}

func (m *Destination) FinalizePull() error {
	dirtyOffsets, err := m.flush()
	if err != nil {
		return err
	}

	if hook := m.hooks.OnAfterFlush; hook != nil {
		if err := hook(dirtyOffsets); err != nil {
			return err
		}
	}

	if m.syncedReadWriter != nil {
		m.syncedReadWriter.MarkAsRemote(dirtyOffsets)
	}

	if m.puller != nil {
		m.puller.FinalizePull(dirtyOffsets)
	}

	return nil
}

func (m *Destination) Close() error {
	if m.syncer != nil {
		_ = m.syncer.Sync()
	}

	if m.dev != nil {
		_ = m.dev.Close()
	}

	if m.puller != nil {
		_ = m.puller.Close()
	}

	if m.serverFile != nil {
		_ = m.serverFile.Close()
	}

	_ = m.close()

	m.wg.Wait()

	if m.errs != nil {
		close(m.errs)

		m.errs = nil
	}

	return nil
}

func (m *Destination) Sync() error {
	return m.syncer.Sync()
}
