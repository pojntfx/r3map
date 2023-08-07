package migration

import (
	"context"
	"errors"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
)

var (
	ErrSeedXORLeech = errors.New("can either seed or leech, but not both at the same time")
)

type MigratorOptions struct {
	ChunkSize    int64
	MaxChunkSize int64

	PullWorkers  int64
	PullPriority func(off int64) int64

	Verbose bool
}

type MigratorHooks struct {
	OnBeforeSync func() error
	OnAfterSync  func(dirtyOffsets []int64) error

	OnBeforeClose func() error

	OnChunkIsLocal func(off int64) error
}

type PathMigrator struct {
	ctx context.Context

	local backend.Backend

	options *MigratorOptions
	hooks   *MigratorHooks

	serverOptions *server.Options
	clientOptions *client.Options

	seeder *PathSeeder

	leecher  *PathLeecher
	released bool

	errs chan error
}

func NewPathMigrator(
	ctx context.Context,

	local backend.Backend,

	options *MigratorOptions,
	hooks *MigratorHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *PathMigrator {
	if options == nil {
		options = &MigratorOptions{}
	}

	if hooks == nil {
		hooks = &MigratorHooks{}
	}

	return &PathMigrator{
		ctx: ctx,

		local: local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (s *PathMigrator) Wait() error {
	for err := range s.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PathMigrator) Seed() (
	path string,
	size int64,
	svc *services.Seeder,
	err error,
) {
	if s.leecher != nil {
		return "", 0, nil, ErrSeedXORLeech
	}

	s.seeder = NewPathSeeder(
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

		close(s.errs)

		s.errs = nil
	}()

	return s.seeder.Open()
}

func (s *PathMigrator) Leech(
	remote *services.SeederRemote,
) (
	finalize func() (
		seed func() (*services.Seeder, error),

		path string,
		err error,
	),

	size int64,
	err error,
) {
	if s.seeder != nil {
		return nil, 0, ErrSeedXORLeech
	}

	s.leecher = NewPathLeecher(
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

		if !s.released {
			close(s.errs)

			s.errs = nil
		}
	}()

	size, err = s.leecher.Open()
	if err != nil {
		return nil, 0, err
	}

	return func() (
		func() (*services.Seeder, error),

		string,
		error,
	) {
		path, err := s.leecher.Finalize()
		if err != nil {
			return nil, "", err
		}

		return func() (
			*services.Seeder,
			error,
		) {
			releasedDev, releasedErrs, releasedWg, releasedDevicePath := s.leecher.Release()

			s.released = true
			if err := s.leecher.Close(); err != nil {
				return nil, err
			}
			s.leecher = nil

			s.seeder = NewPathSeederFromLeecher(
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
				releasedDevicePath,
			)

			go func() {
				if err := s.seeder.Wait(); err != nil {
					s.errs <- err

					return
				}

				close(s.errs)

				s.errs = nil
			}()

			_, _, svc, err := s.seeder.Open()
			if err != nil {
				return nil, err
			}

			return svc, nil
		}, path, err
	}, size, err
}

func (s *PathMigrator) Close() error {
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
