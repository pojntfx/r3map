package migration

import (
	"context"
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
)

type FileLeecher struct {
	path *PathLeecher

	hooks *LeecherHooks

	deviceFile *os.File

	released bool
}

func NewFileLeecher(
	ctx context.Context,

	local backend.Backend,
	remote *services.SeederRemote,

	options *LeecherOptions,
	hooks *LeecherHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileLeecher {
	if hooks == nil {
		hooks = &LeecherHooks{}
	}

	l := &FileLeecher{
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

func (l *FileLeecher) Wait() error {
	return l.path.Wait()
}

func (l *FileLeecher) Open() (*os.File, error) {
	devicePath, _, err := l.path.Open()
	if err != nil {
		return nil, err
	}

	l.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return l.deviceFile, err
}

func (l *FileLeecher) Finalize() error {
	return l.path.Finalize()
}

func (l *FileLeecher) Release() (
	*mount.DirectPathMount,
	chan error,
	*sync.WaitGroup,
	string,
	*os.File,
) {
	l.released = true

	return l.path.Release()
}

func (l *FileLeecher) onBeforeClose() error {
	if hook := l.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}

		l.hooks.OnBeforeClose = nil // Don't call close hook multiple times
	}

	if !l.released && l.deviceFile != nil {
		_ = l.deviceFile.Close()
	}

	return nil
}

func (l *FileLeecher) Close() error {
	return l.path.Close()
}
