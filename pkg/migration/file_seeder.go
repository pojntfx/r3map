package migration

import (
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
)

type FileSeeder struct {
	path *PathSeeder

	hooks *SeederHooks

	deviceFile *os.File
}

func NewFileSeeder(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileSeeder {
	if hooks == nil {
		hooks = &SeederHooks{}
	}

	s := &FileSeeder{
		path: NewPathSeeder(
			local,

			options,
			nil,

			serverOptions,
			clientOptions,
		),

		hooks: hooks,
	}

	s.path.hooks.OnBeforeSync = s.onBeforeSync

	s.path.hooks.OnBeforeClose = s.onBeforeClose

	return s
}

func NewFileSeederFromLeecher(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	dev *mount.DirectPathMount,
	errs chan error,
	wg *sync.WaitGroup,
	devicePath string,
	serverFile *os.File,

	deviceFile *os.File,
) *FileSeeder {
	if hooks == nil {
		hooks = &SeederHooks{}
	}

	s := &FileSeeder{
		path: NewPathSeederFromLeecher(
			local,

			options,
			nil,

			dev,
			errs,
			wg,
			devicePath,
			serverFile,
		),

		hooks: hooks,

		deviceFile: deviceFile,
	}

	s.path.hooks.OnBeforeSync = s.onBeforeSync

	s.path.hooks.OnBeforeClose = s.onBeforeClose

	return s
}

func (s *FileSeeder) Wait() error {
	return s.path.Wait()
}

func (s *FileSeeder) Open() (*os.File, *services.SeederService, error) {
	devicePath, _, svc, err := s.path.Open()
	if err != nil {
		return nil, nil, err
	}

	if s.deviceFile == nil {
		s.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, nil, err
		}
	}

	return s.deviceFile, svc, nil
}

func (s *FileSeeder) onBeforeSync() error {
	if hook := s.hooks.OnBeforeSync; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	if s.deviceFile != nil {
		if err := s.deviceFile.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileSeeder) onBeforeClose() error {
	if hook := s.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}

		s.hooks.OnBeforeClose = nil // Don't call close hook multiple times
	}

	if s.deviceFile != nil {
		_ = s.deviceFile.Close()

		s.deviceFile = nil
	}

	return nil
}

func (s *FileSeeder) Close() error {
	return s.path.Close()
}
