package frontend

import (
	"context"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type SliceFrontend struct {
	path *PathFrontend

	deviceFile *os.File

	slice     []byte
	mmapMount sync.Mutex
}

func NewSliceFrontend(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	options *Options,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceFrontend {
	m := &SliceFrontend{
		path: NewPathFrontend(
			ctx,

			remote,
			local,

			options,
			&Hooks{},

			serverOptions,
			clientOptions,
		),
	}

	m.path.hooks.OnBeforeSync = m.onBeforeSync

	m.path.hooks.OnBeforeClose = m.onBeforeClose
	m.path.hooks.OnAfterClose = m.onAfterClose

	return m
}

func (m *SliceFrontend) Wait() error {
	return m.path.Wait()
}

func (m *SliceFrontend) Open() ([]byte, error) {
	devicePath, size, err := m.path.Open()
	if err != nil {
		return []byte{}, err
	}

	m.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return []byte{}, err
	}

	m.slice, err = syscall.Mmap(
		int(m.deviceFile.Fd()),
		0,
		int(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return []byte{}, err
	}

	return m.slice, nil
}

func (m *SliceFrontend) onBeforeSync() error {
	m.mmapMount.Lock()
	if m.path != nil {
		if _, _, err := syscall.Syscall(
			syscall.SYS_MSYNC,
			uintptr(unsafe.Pointer(&m.slice[0])),
			uintptr(len(m.slice)),
			uintptr(syscall.MS_SYNC),
		); err != 0 {
			m.mmapMount.Unlock()

			return err
		}
	}
	m.mmapMount.Unlock()

	return nil
}

func (m *SliceFrontend) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()
	}

	return nil
}

func (m *SliceFrontend) onAfterClose() error {
	m.mmapMount.Lock()
	if m.slice != nil {
		_ = syscall.Munmap(m.slice)

		m.slice = nil
	}
	m.mmapMount.Unlock()

	return nil
}

func (m *SliceFrontend) Close() error {
	return m.path.Close()
}

func (m *SliceFrontend) Sync() error {
	return m.path.Sync()
}
