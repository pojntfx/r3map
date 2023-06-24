package migration

import (
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type FileSeederHooks struct {
	OnBeforeSync func() error

	OnAfterClose func() error
}

type FileSeeder struct {
	path *PathSeeder

	deviceFile *os.File
}

func NewFileSeeder(
	local backend.Backend,

	options *SeederOptions,
	hooks *FileSeederHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileSeeder {
	h := &SeederHooks{}
	if hooks != nil {
		h.OnBeforeSync = hooks.OnBeforeSync
		h.OnAfterClose = hooks.OnAfterClose
	}

	s := &FileSeeder{
		path: NewPathSeeder(
			local,

			options,
			h,

			serverOptions,
			clientOptions,
		),
	}

	s.path.hooks.OnBeforeClose = s.onBeforeClose

	return s
}

func (s *FileSeeder) Wait() error {
	return s.path.Wait()
}

func (s *FileSeeder) Open() (*os.File, *services.Seeder, error) {
	devicePath, _, svc, err := s.path.Open()
	if err != nil {
		return nil, nil, err
	}

	s.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	return s.deviceFile, svc, nil
}

func (s *FileSeeder) onBeforeClose() error {
	if s.deviceFile != nil {
		_ = s.deviceFile.Close()
	}

	return nil
}

func (s *FileSeeder) Close() error {
	return s.path.Close()
}
