package mount

import (
	"context"
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type ManagedSliceMountHooks struct {
	OnChunkIsLocal func(off int64) error
}

type ManagedSliceMount struct {
	path *ManagedPathMount

	deviceFile *os.File

	slice     mmap.MMap
	mmapMount sync.Mutex
}

func NewManagedSliceMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	options *ManagedMountOptions,
	hooks *ManagedSliceMountHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *ManagedSliceMount {
	h := &ManagedMountHooks{}
	if hooks != nil {
		h.OnChunkIsLocal = hooks.OnChunkIsLocal
	}

	m := &ManagedSliceMount{
		path: NewManagedPathMount(
			ctx,

			remote,
			local,

			options,
			h,

			serverOptions,
			clientOptions,
		),
	}

	m.path.hooks.OnBeforeSync = m.onBeforeSync

	m.path.hooks.OnBeforeClose = m.onBeforeClose

	return m
}

func (m *ManagedSliceMount) Wait() error {
	return m.path.Wait()
}

func (m *ManagedSliceMount) Open() ([]byte, error) {
	devicePath, size, err := m.path.Open()
	if err != nil {
		return []byte{}, err
	}

	m.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return []byte{}, err
	}

	m.slice, err = mmap.MapRegion(m.deviceFile, int(size), mmap.RDWR, 0, 0)
	if err != nil {
		return []byte{}, err
	}

	// We _MUST_ lock this slice so that it does not get paged out
	// If it does, the Go GC tries to manage it, deadlocking _the entire runtime_
	if err := m.slice.Lock(); err != nil {
		return []byte{}, err
	}

	return m.slice, nil
}

func (m *ManagedSliceMount) onBeforeSync() error {
	m.mmapMount.Lock()
	if m.slice != nil {
		if err := m.slice.Flush(); err != nil {
			return err
		}
	}
	m.mmapMount.Unlock()

	return nil
}

func (m *ManagedSliceMount) onBeforeClose() error {
	m.mmapMount.Lock()
	if m.slice != nil {
		_ = m.slice.Unlock()

		_ = m.slice.Unmap()

		m.slice = nil
	}
	m.mmapMount.Unlock()

	if m.deviceFile != nil {
		_ = m.deviceFile.Close()
	}

	return nil
}

func (m *ManagedSliceMount) Close() error {
	return m.path.Close()
}

func (m *ManagedSliceMount) Sync() error {
	if err := m.onBeforeSync(); err != nil {
		return err
	}

	return m.path.Sync()
}
