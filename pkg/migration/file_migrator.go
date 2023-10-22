package migration

import (
	"context"
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type FileMigrator struct {
	ctx context.Context

	local backend.Backend

	options *MigratorOptions
	hooks   *MigratorHooks

	serverOptions *server.Options
	clientOptions *client.Options

	seeder *FileSeeder

	leecher  *FileLeecher
	released bool

	closeLock sync.Mutex

	errs chan error
}

func NewFileMigrator(
	ctx context.Context,

	local backend.Backend,

	options *MigratorOptions,
	hooks *MigratorHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileMigrator {
	if options == nil {
		options = &MigratorOptions{}
	}

	if hooks == nil {
		hooks = &MigratorHooks{}
	}

	return &FileMigrator{
		ctx: ctx,

		local: local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (s *FileMigrator) Wait() error {
	for err := range s.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *FileMigrator) Seed() (
	file *os.File,
	svc *services.SeederService,
	err error,
) {
	if s.leecher != nil {
		return nil, nil, ErrSeedXORLeech
	}

	s.seeder = NewFileSeeder(
		s.local,

		&SeederOptions{
			ChunkSize:    s.options.ChunkSize,
			MaxChunkSize: s.options.MaxChunkSize,

			Verbose: s.options.Verbose,
		},
		&SeederHooks{
			OnBeforeSync: s.hooks.OnBeforeSync,

			OnBeforeClose: s.hooks.OnBeforeClose,
		},

		s.serverOptions,
		s.clientOptions,
	)

	go func() {
		if err := s.seeder.Wait(); err != nil {
			s.errs <- err

			return
		}

		if s.errs != nil {
			close(s.errs)

			s.errs = nil
		}
	}()

	return s.seeder.Open()
}

func (s *FileMigrator) Leech(
	remote *services.SeederRemote,
) (
	finalize func() (
		seed func() (
			svc *services.SeederService,
			err error,
		),

		file *os.File,
		err error,
	),

	err error,
) {
	if s.seeder != nil {
		return nil, ErrSeedXORLeech
	}

	s.leecher = NewFileLeecher(
		s.ctx,

		s.local,
		remote,

		&LeecherOptions{
			ChunkSize: s.options.ChunkSize,

			PullWorkers:  s.options.PullWorkers,
			PullPriority: s.options.PullPriority,

			Verbose: s.options.Verbose,
		},
		&LeecherHooks{
			OnAfterSync: s.hooks.OnAfterSync,

			OnBeforeClose: s.hooks.OnBeforeClose,

			OnChunkIsLocal: s.hooks.OnChunkIsLocal,
		},

		nil,
		nil,
	)

	go func() {
		if err := s.leecher.Wait(); err != nil {
			s.errs <- err

			return
		}

		if !s.released && s.errs != nil {
			close(s.errs)

			s.errs = nil
		}
	}()

	if err = s.leecher.Open(); err != nil {
		return nil, err
	}

	return func() (
		func() (*services.SeederService, error),

		*os.File,
		error,
	) {
		file, err := s.leecher.Finalize()
		if err != nil {
			return nil, nil, err
		}

		return func() (
			*services.SeederService,
			error,
		) {
			releasedDev, releasedErrs, releasedWg, releasedDeviceFile, releasedServerFile := s.leecher.Release()

			s.released = true
			if err := s.leecher.Close(); err != nil {
				return nil, err
			}
			s.leecher = nil

			s.seeder = NewFileSeederFromLeecher(
				s.local,

				&SeederOptions{
					ChunkSize:    s.options.ChunkSize,
					MaxChunkSize: s.options.MaxChunkSize,

					Verbose: s.options.Verbose,
				},
				&SeederHooks{
					OnBeforeSync: s.hooks.OnBeforeSync,

					OnBeforeClose: s.hooks.OnBeforeClose,
				},

				releasedDev,
				releasedErrs,
				releasedWg,
				releasedDeviceFile,
				releasedServerFile,

				file,
			)

			go func() {
				if err := s.seeder.Wait(); err != nil {
					s.errs <- err

					return
				}

				if s.errs != nil {
					close(s.errs)

					s.errs = nil
				}
			}()

			_, svc, err := s.seeder.Open()
			if err != nil {
				return nil, err
			}

			return svc, nil
		}, file, err
	}, err
}

func (s *FileMigrator) Close() error {
	s.closeLock.Lock()
	defer s.closeLock.Unlock()

	if s.seeder != nil {
		_ = s.seeder.Close()
	}

	if s.leecher != nil {
		_ = s.leecher.Close()
	}

	if s.errs != nil {
		close(s.errs)

		s.errs = nil
	}

	return nil
}
