package migration

import (
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	bbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

type SeederOptions struct {
	ChunkSize int64

	Verbose bool
}

type SeederHooks struct {
	OnBeforeFlush func() error
}

type Seeder struct {
	local backend.Backend

	options *SeederOptions
	hooks   *SeederHooks

	serverOptions *server.Options
	clientOptions *client.Options

	wg   sync.WaitGroup
	errs chan error
}

func NewSeeder(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *Seeder {
	if options == nil {
		options = &SeederOptions{}
	}

	if options.ChunkSize <= 0 {
		options.ChunkSize = 4096
	}

	if hooks == nil {
		hooks = &SeederHooks{}
	}

	return &Seeder{
		local: local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (s *Seeder) Open() (string, int64, *services.Source, error) {
	size, err := s.local.Size()
	if err != nil {
		return "", 0, nil, err
	}

	devicePath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		return "", 0, nil, err
	}

	serverFile, err := os.Open(devicePath)
	if err != nil {
		return "", 0, nil, err
	}
	defer serverFile.Close()

	tr := chunks.NewTrackingReadWriterAt(s.local)

	b := bbackend.NewReaderAtBackend(
		chunks.NewArbitraryReadWriterAt(
			chunks.NewChunkedReadWriterAt(
				tr,
				s.options.ChunkSize,
				size/s.options.ChunkSize,
			),
			s.options.ChunkSize,
		),
		s.local.Size,
		s.local.Sync,
		false,
	)

	dev := device.NewDevice(
		b,
		serverFile,

		s.serverOptions,
		s.clientOptions,
	)
	defer dev.Close()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if err := dev.Wait(); err != nil {
			s.errs <- err

			return
		}
	}()

	if err := dev.Open(); err != nil {
		return "", 0, nil, err
	}

	flushed := false
	return "", 0, services.NewSource(
			b,
			s.options.Verbose,
			func() error {
				tr.Track()

				return nil
			},
			func() ([]int64, error) {
				if hook := s.hooks.OnBeforeFlush; hook != nil {
					if err := hook(); err != nil {
						return []int64{}, err
					}
				}

				// TODO: `nil` these ressources/integrate with `s.Close()`
				if err := dev.Close(); err != nil {
					return []int64{}, err
				}

				if err := serverFile.Close(); err != nil {
					return []int64{}, err
				}

				rv := tr.Flush()

				flushed = true

				return rv, nil
			},
			func() error {
				// Stop seeding
				if flushed {
					// TODO: `return s.Close()` here once implemented
				}

				return nil
			},
		),
		nil
}
