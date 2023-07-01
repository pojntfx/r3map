package mount

import (
	"context"
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type ManagedFileMount struct {
	path *ManagedPathMount

	deviceFile *os.File
}

type ManagedFileMountHooks struct {
	OnChunkIsLocal func(off int64) error
}

func NewManagedFileMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	options *ManagedMountOptions,
	hooks *ManagedFileMountHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *ManagedFileMount {
	h := &ManagedMountHooks{}
	if hooks != nil {
		h.OnChunkIsLocal = hooks.OnChunkIsLocal
	}

	m := &ManagedFileMount{
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

func (m *ManagedFileMount) Wait() error {
	return m.path.Wait()
}

func (m *ManagedFileMount) Open() (*os.File, error) {
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

func (m *ManagedFileMount) onBeforeSync() error {
	if m.deviceFile != nil {
		if err := m.deviceFile.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (m *ManagedFileMount) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()

		m.deviceFile = nil
	}

	return nil
}

func (m *ManagedFileMount) Close() error {
	return m.path.Close()
}

func (m *ManagedFileMount) Sync() error {
	if err := m.onBeforeSync(); err != nil {
		return err
	}

	return m.path.Sync()
}
