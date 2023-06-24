package migration

import (
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type SliceSeeder struct {
	path *PathSeeder

	deviceFile *os.File

	slice     mmap.MMap
	mmapMount sync.Mutex
}

func NewSliceSeeder(
	local backend.Backend,

	options *SeederOptions,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceSeeder {
	h := &SeederHooks{}

	s := &SliceSeeder{
		path: NewPathSeeder(
			local,

			options,
			h,

			serverOptions,
			clientOptions,
		),
	}

	s.path.hooks.OnBeforeSync = s.onBeforeSync

	s.path.hooks.OnBeforeClose = s.onBeforeClose
	s.path.hooks.OnAfterClose = s.onAfterClose

	return s
}

func (s *SliceSeeder) Wait() error {
	return s.path.Wait()
}

func (s *SliceSeeder) Open() ([]byte, *services.Seeder, error) {
	devicePath, size, svc, err := s.path.Open()
	if err != nil {
		return []byte{}, nil, err
	}

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

	return s.slice, svc, nil
}

func (m *SliceSeeder) onBeforeSync() error {
	m.mmapMount.Lock()
	if m.slice != nil {
		if err := m.slice.Flush(); err != nil {
			return err
		}
	}
	m.mmapMount.Unlock()

	return nil
}

func (m *SliceSeeder) onBeforeClose() error {
	if m.deviceFile != nil {
		_ = m.deviceFile.Close()
	}

	return nil
}

func (m *SliceSeeder) onAfterClose() error {
	m.mmapMount.Lock()
	if m.slice != nil {
		_ = m.slice.Unlock()

		_ = m.slice.Unmap()

		m.slice = nil
	}
	m.mmapMount.Unlock()

	return nil
}

func (s *SliceSeeder) Close() error {
	return s.path.Close()
}
