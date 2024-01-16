package migration

import (
	"context"
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
)

type SliceLeecher struct {
	path *PathLeecher

	hooks *LeecherHooks

	deviceFile *os.File

	slice      mmap.MMap
	mmapMount  sync.Mutex
	size       int64
	devicePath string

	released bool
}

func NewSliceLeecher(
	ctx context.Context,

	local backend.Backend,
	remote *services.SeederRemote,

	options *LeecherOptions,
	hooks *LeecherHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceLeecher {
	if hooks == nil {
		hooks = &LeecherHooks{}
	}

	l := &SliceLeecher{
		path: NewPathLeecher(
			ctx,

			local,
			remote,

			options,
			nil,

			serverOptions,
			clientOptions,
		),

		hooks: hooks,
	}

	l.path.hooks.OnAfterSync = hooks.OnAfterSync

	l.path.hooks.OnBeforeClose = l.onBeforeClose

	l.path.hooks.OnChunkIsLocal = hooks.OnChunkIsLocal

	return l
}

func (l *SliceLeecher) Wait() error {
	return l.path.Wait()
}

func (l *SliceLeecher) Open() error {
	devicePath, size, err := l.path.Open()
	if err != nil {
		return err
	}

	l.size = size
	l.devicePath = devicePath

	return nil
}

func (l *SliceLeecher) Finalize() ([]byte, error) {
	if err := l.path.Finalize(); err != nil {
		return nil, err
	}

	var err error
	l.deviceFile, err = os.OpenFile(l.devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	l.slice, err = mmap.MapRegion(l.deviceFile, int(l.size), mmap.RDWR, 0, 0)
	if err != nil {
		return nil, err
	}

	// We _MUST_ lock this slice so that it does not get paged out
	// If it does, the Go GC tries to manage it, deadlocking _the entire runtime_
	if err := l.slice.Lock(); err != nil {
		return nil, err
	}

	return l.slice, nil
}

func (l *SliceLeecher) Release() (
	*mount.DirectPathMount,
	chan error,
	*sync.WaitGroup,
	string,
	*os.File,
) {
	l.released = true

	return l.path.Release()
}

func (l *SliceLeecher) onBeforeClose() error {
	if hook := l.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}

		l.hooks.OnBeforeClose = nil // Don't call close hook multiple times
	}

	if !l.released {
		l.mmapMount.Lock()
		if l.slice != nil {
			_ = l.slice.Unlock()

			_ = l.slice.Unmap()

			l.slice = nil
		}
		l.mmapMount.Unlock()

		if l.deviceFile != nil {
			_ = l.deviceFile.Close()
		}
	}

	return nil
}

func (l *SliceLeecher) Close() error {
	return l.path.Close()
}
