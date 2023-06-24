package migration

import (
	"context"
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type SliceLeecherHooks struct {
	OnAfterSync func(dirtyOffsets []int64) error

	OnChunkIsLocal func(off int64) error
}

type SliceLeecher struct {
	path *PathLeecher

	deviceFile *os.File

	slice     mmap.MMap
	mmapMount sync.Mutex
}

func NewSliceLeecher(
	ctx context.Context,

	local backend.Backend,
	remote *services.SeederRemote,

	options *LeecherOptions,
	hooks *SliceLeecherHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceLeecher {
	h := &LeecherHooks{}
	if hooks != nil {
		h.OnAfterSync = hooks.OnAfterSync

		h.OnChunkIsLocal = hooks.OnChunkIsLocal
	}

	l := &SliceLeecher{
		path: NewPathLeecher(
			ctx,

			local,
			remote,

			options,
			h,

			serverOptions,
			clientOptions,
		),
	}

	l.path.hooks.OnBeforeClose = l.onBeforeClose
	l.path.hooks.OnAfterClose = l.onAfterClose

	return l
}

func (l *SliceLeecher) Wait() error {
	return l.path.Wait()
}

// Do not read or write from the returned slice before `Finalize()` has been called
// If you read before `Finalize()`, you'll get incomplete data but no corruption
// If you write before `Finalize()`, you'll corrupt the received data
func (l *SliceLeecher) Open() ([]byte, error) {
	devicePath, size, err := l.path.Open()
	if err != nil {
		return []byte{}, err
	}

	l.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return []byte{}, err
	}

	l.slice, err = mmap.MapRegion(l.deviceFile, int(size), mmap.RDWR, 0, 0)
	if err != nil {
		return []byte{}, err
	}

	// We _MUST_ lock this slice so that it does not get paged out
	// If it does, the Go GC tries to manage it, deadlocking _the entire runtime_
	if err := l.slice.Lock(); err != nil {
		return []byte{}, err
	}

	return l.slice, nil
}

func (l *SliceLeecher) Finalize() error {
	return l.path.Finalize()
}

func (l *SliceLeecher) onBeforeClose() error {
	if l.deviceFile != nil {
		_ = l.deviceFile.Close()
	}

	return nil
}

func (l *SliceLeecher) onAfterClose() error {
	l.mmapMount.Lock()
	if l.slice != nil {
		_ = l.slice.Unlock()

		_ = l.slice.Unmap()

		l.slice = nil
	}
	l.mmapMount.Unlock()

	return nil
}

func (l *SliceLeecher) Close() error {
	return l.path.Close()
}
