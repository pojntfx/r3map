package migration

import (
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	bbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

type SeederOptions struct {
	ChunkSize    int64
	MaxChunkSize int64

	Verbose bool
}

type SeederHooks struct {
	OnBeforeSync func() error

	OnBeforeClose func() error
}

type PathSeeder struct {
	local backend.Backend

	options *SeederOptions
	hooks   *SeederHooks

	serverOptions *server.Options
	clientOptions *client.Options

	serverFile *os.File
	dev        *mount.DirectPathMount

	devicePath string

	wg   *sync.WaitGroup
	errs chan error
}

func NewPathSeeder(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *PathSeeder {
	if options == nil {
		options = &SeederOptions{}
	}

	if options.ChunkSize <= 0 {
		options.ChunkSize = client.MaximumBlockSize
	}

	if hooks == nil {
		hooks = &SeederHooks{}
	}

	return &PathSeeder{
		local: local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		wg:   &sync.WaitGroup{},
		errs: make(chan error),
	}
}

func NewPathSeederFromLeecher(
	local backend.Backend,

	options *SeederOptions,
	hooks *SeederHooks,

	dev *mount.DirectPathMount,
	errs chan error,
	wg *sync.WaitGroup,
	devicePath string,
	serverFile *os.File,
) *PathSeeder {
	if options == nil {
		options = &SeederOptions{}
	}

	if options.ChunkSize <= 0 {
		options.ChunkSize = client.MaximumBlockSize
	}

	if hooks == nil {
		hooks = &SeederHooks{}
	}

	return &PathSeeder{
		local: local,

		options: options,
		hooks:   hooks,

		dev:        dev,
		errs:       errs,
		wg:         wg,
		devicePath: devicePath,
		serverFile: serverFile,
	}
}

func (s *PathSeeder) Wait() error {
	for err := range s.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PathSeeder) Open() (string, int64, *services.SeederService, error) {
	size, err := s.local.Size()
	if err != nil {
		return "", 0, nil, err
	}

	tr := chunks.NewTrackingReadWriterAt(s.local)

	b := bbackend.NewReaderAtBackend(
		chunks.NewArbitraryReadWriterAt(
			tr,
			s.options.ChunkSize,
		),
		s.local.Size,
		s.local.Sync,
		false,
	)

	if s.dev == nil {
		s.devicePath, err = utils.FindUnusedNBDDevice()
		if err != nil {
			return "", 0, nil, err
		}

		s.serverFile, err = os.Open(s.devicePath)
		if err != nil {
			return "", 0, nil, err
		}

		s.dev = mount.NewDirectPathMount(
			b,
			s.serverFile,

			s.serverOptions,
			s.clientOptions,
		)

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			if err := s.dev.Wait(); err != nil {
				s.errs <- err

				return
			}
		}()

		if err := s.dev.Open(); err != nil {
			return "", 0, nil, err
		}
	} else {
		s.dev.SwapBackend(b)
	}

	synced := false
	return s.devicePath, size, services.NewSeederService(
			b,
			s.options.Verbose,
			func() error {
				tr.Track()

				return nil
			},
			func() ([]int64, error) {
				if hook := s.hooks.OnBeforeSync; hook != nil {
					if err := hook(); err != nil {
						return []int64{}, err
					}
				}

				rv := tr.Sync()

				synced = true

				return rv, nil
			},
			func() error {
				// Stop seeding
				if synced {
					return s.Close()
				}

				return nil
			},
			s.options.MaxChunkSize,
		),
		nil
}

func (s *PathSeeder) Close() error {
	if hook := s.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}

		s.hooks.OnBeforeClose = nil // Don't call close hook multiple times
	}

	if s.dev != nil {
		_ = s.dev.Close()
	}

	if s.serverFile != nil {
		_ = s.serverFile.Close()
	}

	s.wg.Wait()

	if s.errs != nil {
		close(s.errs)

		s.errs = nil
	}

	return nil
}
