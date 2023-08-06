package migration

import (
	"context"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/r3map/pkg/services"
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

func (s *PathMigrator) Seed() (string, int64, *services.Seeder, error) {
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

func (s *PathMigrator) Leech(remote *services.SeederRemote) (string, int64, *services.Seeder, error) {
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
		if err := s.seeder.Wait(); err != nil {
			s.errs <- err

			return
		}

		if !s.released {
			close(s.errs)

			s.errs = nil
		}
	}()

	return s.seeder.Open()
}

func (s *PathMigrator) Close() error {
	if s.seeder != nil {
		_ = s.seeder.Close()
	}

	if s.errs != nil {
		close(s.errs)

		s.errs = nil
	}

	return nil
}
