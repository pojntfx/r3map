package device

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

type SliceMount struct {
	mount *Mount

	deviceFile *os.File

	slice     []byte
	mmapMount sync.Mutex
}

func NewSliceMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	mountOptions *MountOptions,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceMount {
	m := &SliceMount{
		mount: NewMount(
			ctx,

			remote,
			local,

			mountOptions,
			&MountHooks{},

			serverOptions,
			clientOptions,
		),
	}

	m.mount.mountHooks.OnBeforeSync = m.onBeforeSync

	m.mount.mountHooks.OnBeforeClose = m.onBeforeClose
	m.mount.mountHooks.OnAfterClose = m.onAfterClose

	return m
}

func (m *SliceMount) Wait() error {
	return m.mount.Wait()
}

func (m *SliceMount) Open() ([]byte, error) {
	devicePath, size, err := m.mount.Open()
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

func (m *SliceMount) onBeforeSync() error {
	m.mmapMount.Lock()
	if m.mount != nil {
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

func (m *SliceMount) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()
	}

	return nil
}

func (m *SliceMount) onAfterClose() error {
	m.mmapMount.Lock()
	if m.slice != nil {
		_ = syscall.Munmap(m.slice)

		m.slice = nil
	}
	m.mmapMount.Unlock()

	return nil
}

func (m *SliceMount) Close() error {
	return m.mount.Close()
}

func (m *SliceMount) Sync() error {
	return m.mount.Sync()
}
