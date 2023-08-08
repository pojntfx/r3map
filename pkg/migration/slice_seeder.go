package migration

import (
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
)

type SliceSeeder struct {
	path *PathSeeder

	hooks *SeederHooks

	deviceFile *os.File

	slice     mmap.MMap
	mmapMount sync.Mutex
}

func NewSliceSeeder(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceSeeder {
	if hooks == nil {
		hooks = &SeederHooks{}
	}

	s := &SliceSeeder{
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

func NewSliceSeederFromLeecher(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	dev *mount.DirectPathMount,
	errs chan error,
	wg *sync.WaitGroup,
	devicePath string,

	deviceSlice []byte,
) *SliceSeeder {
	if hooks == nil {
		hooks = &SeederHooks{}
	}

	s := &SliceSeeder{
		path: NewPathSeederFromLeecher(
			local,

			options,
			nil,

			dev,
			errs,
			wg,
			devicePath,
		),

		hooks: hooks,

		slice: deviceSlice,
	}

	s.path.hooks.OnBeforeSync = s.onBeforeSync

	s.path.hooks.OnBeforeClose = s.onBeforeClose

	return s
}

func (s *SliceSeeder) Wait() error {
	return s.path.Wait()
}

func (s *SliceSeeder) Open() ([]byte, *services.SeederService, error) {
	devicePath, size, svc, err := s.path.Open()
	if err != nil {
		return []byte{}, nil, err
	}

	if s.deviceFile == nil {
		s.deviceFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
		if err != nil {
			return []byte{}, nil, err
		}

		s.slice, err = mmap.MapRegion(s.deviceFile, int(size), mmap.RDWR, 0, 0)
		if err != nil {
			return []byte{}, nil, err
		}

		// We _MUST_ lock this slice so that it does not get paged out
		// If it does, the Go GC tries to manage it, deadlocking _the entire runtime_
		if err := s.slice.Lock(); err != nil {
			return []byte{}, nil, err
		}
	}

	return s.slice, svc, nil
}

func (s *SliceSeeder) onBeforeSync() error {
	if hook := s.hooks.OnBeforeSync; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	s.mmapMount.Lock()
	if s.slice != nil {
		if err := s.slice.Flush(); err != nil {
			return err
		}
	}
	s.mmapMount.Unlock()

	return nil
}

func (s *SliceSeeder) onBeforeClose() error {
	if hook := s.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	s.mmapMount.Lock()
	if s.slice != nil {
		_ = s.slice.Unlock()

		_ = s.slice.Unmap()

		s.slice = nil
	}
	s.mmapMount.Unlock()

	if s.deviceFile != nil {
		_ = s.deviceFile.Close()
	}

	return nil
}

func (s *SliceSeeder) Close() error {
	return s.path.Close()
}
