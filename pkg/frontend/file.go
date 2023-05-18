package frontend

import (
	"context"
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type FileFrontend struct {
	path *PathFrontend

	deviceFile *os.File
}

func NewFileFrontend(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	options *Options,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileFrontend {
	m := &FileFrontend{
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

	return m
}

func (m *FileFrontend) Wait() error {
	return m.path.Wait()
}

func (m *FileFrontend) Open() (*os.File, error) {
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

func (m *FileFrontend) onBeforeSync() error {
	if m.deviceFile != nil {
		if err := m.deviceFile.Sync(); err != nil {
			return nil
		}
	}

	return nil
}

func (m *FileFrontend) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()
	}

	return nil
}

func (m *FileFrontend) Close() error {
	return m.path.Close()
}

func (m *FileFrontend) Sync() error {
	return m.path.Sync()
}
