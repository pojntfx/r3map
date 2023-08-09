package migration

import (
	"context"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

type SliceMigrator struct {
	ctx context.Context

	local backend.Backend

	options *MigratorOptions
	hooks   *MigratorHooks

	serverOptions *server.Options
	clientOptions *client.Options

	seeder *SliceSeeder

	leecher  *SliceLeecher
	released bool

	errs chan error
}

func NewSliceMigrator(
	ctx context.Context,

	local backend.Backend,

	options *MigratorOptions,
	hooks *MigratorHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *SliceMigrator {
	if options == nil {
		options = &MigratorOptions{}
	}

	if hooks == nil {
		hooks = &MigratorHooks{}
	}

	return &SliceMigrator{
		ctx: ctx,

		local: local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (s *SliceMigrator) Wait() error {
	for err := range s.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SliceMigrator) Seed() (
	slice []byte,
	svc *services.SeederService,
	err error,
) {
	if s.leecher != nil {
		return nil, nil, ErrSeedXORLeech
	}

	s.seeder = NewSliceSeeder(
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

func (s *SliceMigrator) Leech(
	remote *services.SeederRemote,
) (
	finalize func() (
		seed func() (
			svc *services.SeederService,
			err error,
		),

		slice []byte,
		err error,
	),

	err error,
) {
	if s.seeder != nil {
		return nil, ErrSeedXORLeech
	}

	s.leecher = NewSliceLeecher(
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

		[]byte,
		error,
	) {
		slice, err := s.leecher.Finalize()
		if err != nil {
			return nil, nil, err
		}

		return func() (
			*services.SeederService,
			error,
		) {
			releasedDev, releasedErrs, releasedWg, releasedDeviceSlice := s.leecher.Release()

			s.released = true
			if err := s.leecher.Close(); err != nil {
				return nil, err
			}
			s.leecher = nil

			s.seeder = NewSliceSeederFromLeecher(
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
				releasedDeviceSlice,

				slice,
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
		}, slice, err
	}, err
}

func (s *SliceMigrator) Close() error {
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
