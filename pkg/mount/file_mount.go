package mount

import (
	"context"
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type FileMount struct {
	path *PathMount

	deviceFile *os.File
}

type FileMountHooks struct {
	OnChunkIsLocal func(off int64) error
}

func NewFileMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	options *MountOptions,
	hooks *FileMountHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileMount {
	h := &MountHooks{}
	if hooks != nil {
		h.OnChunkIsLocal = hooks.OnChunkIsLocal
	}

	m := &FileMount{
		path: NewPathMount(
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

func (m *FileMount) Wait() error {
	return m.path.Wait()
}

func (m *FileMount) Open() (*os.File, error) {
	devicePath, _, err := m.path.Open()
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
			return err
		}
	}

	return nil
}

func (m *FileMount) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()

		m.deviceFile = nil
	}

	return nil
}

func (m *FileMount) Close() error {
	return m.path.Close()
}

func (m *FileMount) Sync() error {
	if err := m.onBeforeSync(); err != nil {
		return err
	}

	return m.path.Sync()
}
