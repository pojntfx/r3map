package device

import (
	"context"
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type FileMount struct {
	mount *Mount

	deviceFile *os.File
}

func NewFileMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	mountOptions *MountOptions,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileMount {
	m := &FileMount{
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

	return m
}

func (m *FileMount) Wait() error {
	return m.mount.Wait()
}

func (m *FileMount) Open() (*os.File, error) {
	devicePath, _, err := m.mount.Open()
	if err != nil {
		return nil, err
	}

	m.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return m.deviceFile, nil
}

func (m *FileMount) onBeforeSync() error {
	if m.deviceFile != nil {
		if err := m.deviceFile.Sync(); err != nil {
			return nil
		}
	}

	return nil
}

func (m *FileMount) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()
	}

	return nil
}

func (m *FileMount) Close() error {
	return m.mount.Close()
}

func (m *FileMount) Sync() error {
	return m.mount.Sync()
}
