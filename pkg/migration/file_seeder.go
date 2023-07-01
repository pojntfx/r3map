package migration

import (
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type FileSeeder struct {
	path *PathSeeder

	deviceFile *os.File
}

func NewFileSeeder(
	local backend.Backend,

	options *SeederOptions,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileSeeder {
	s := &FileSeeder{
		path: NewPathSeeder(
			local,

			options,
			nil,

			serverOptions,
			clientOptions,
		),
	}

	s.path.hooks.OnBeforeSync = s.onBeforeSync

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

func (s *FileSeeder) onBeforeSync() error {
	if s.deviceFile != nil {
		if err := s.deviceFile.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileSeeder) onBeforeClose() error {
	if s.deviceFile != nil {
		_ = s.deviceFile.Close()

		s.deviceFile = nil
	}

	return nil
}

func (s *FileSeeder) Close() error {
	return s.path.Close()
}
