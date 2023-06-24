package migration

import (
	"context"
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type FileLeecherHooks struct {
	OnAfterSync func(dirtyOffsets []int64) error

	OnAfterClose func() error

	OnChunkIsLocal func(off int64) error
}

type FileLeecher struct {
	path *PathLeecher

	deviceFile *os.File
}

func NewFileLeecher(
	ctx context.Context,

	local backend.Backend,
	remote *services.SeederRemote,

	options *LeecherOptions,
	hooks *FileLeecherHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileLeecher {
	h := &LeecherHooks{}
	if hooks != nil {
		h.OnAfterSync = hooks.OnAfterSync

		h.OnAfterClose = hooks.OnAfterClose

		h.OnChunkIsLocal = hooks.OnChunkIsLocal
	}

	l := &FileLeecher{
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

	return l
}

func (l *FileLeecher) Wait() error {
	return l.path.Wait()
}

func (l *FileLeecher) Open() error {
	_, err := l.path.Open()

	return err
}

func (l *FileLeecher) Finalize() (*os.File, error) {
	devicePath, err := l.path.Finalize()
	if err != nil {
		return nil, err
	}

	l.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return l.deviceFile, nil
}

func (l *FileLeecher) onBeforeClose() error {
	if l.deviceFile != nil {
		_ = l.deviceFile.Close()
	}

	return nil
}

func (l *FileLeecher) Close() error {
	return l.path.Close()
}
