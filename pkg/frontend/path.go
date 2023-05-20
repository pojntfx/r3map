package frontend

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
	bbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/utils"
)

type Options struct {
	ChunkSize int64

	PullWorkers int64
	PullFirst   bool

	PushWorkers  int64
	PushInterval time.Duration

	Verbose bool
}

type Hooks struct {
	OnBeforeSync func() error

	OnBeforeClose func() error
	OnAfterClose  func() error

	OnChunkIsLocal func(off int64) error
}

type PathFrontend struct {
	ctx context.Context

	remote,
	local,
	syncer backend.Backend

	options *Options
	hooks   *Hooks

	serverOptions *server.Options
	clientOptions *client.Options

	serverFile *os.File
	pusher     *chunks.Pusher
	puller     *chunks.Puller
	dev        *device.Device

	wg   sync.WaitGroup
	errs chan error
}

func NewPathFrontend(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	options *Options,
	hooks *Hooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *PathFrontend {
	if options == nil {
		options = &Options{}
	}

	if options.ChunkSize <= 0 {
		options.ChunkSize = 4096
	}

	if options.PushInterval == 0 {
		options.PushInterval = time.Second * 20
	}

	if hooks == nil {
		hooks = &Hooks{}
	}

	return &PathFrontend{
		ctx: ctx,

		remote: remote,
		local:  local,

		options: options,
		hooks:   hooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (m *PathFrontend) Wait() error {
	for err := range m.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *PathFrontend) Open() (string, int64, error) {
	size, err := m.remote.Size()
	if err != nil {
		return "", 0, err
	}
	chunkCount := size / m.options.ChunkSize

	devicePath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		return "", 0, err
	}

	m.serverFile, err = os.Open(devicePath)
	if err != nil {
		return "", 0, err
	}

	var local chunks.ReadWriterAt
	if m.options.PushWorkers > 0 {
		m.pusher = chunks.NewPusher(
			m.ctx,
			chunks.NewChunkedReadWriterAt(m.local, m.options.ChunkSize, chunkCount),
			m.remote,
			m.options.ChunkSize,
			m.options.PushInterval,
		)

		m.wg.Add(1)
		go func() {
			defer m.wg.Done()

			if err := m.pusher.Wait(); err != nil {
				m.errs <- err

				return
			}
		}()

		if err := m.pusher.Open(m.options.PushWorkers); err != nil {
			return "", 0, err
		}

		local = m.pusher
	} else {
		local = chunks.NewChunkedReadWriterAt(m.local, m.options.ChunkSize, chunkCount)
	}

	syncedReadWriter := chunks.NewSyncedReadWriterAt(m.remote, local, func(off int64) error {
		if m.options.PushWorkers > 0 {
			if err := local.(*chunks.Pusher).MarkOffsetPushable(off); err != nil {
				return err
			}

			if hook := m.hooks.OnChunkIsLocal; hook != nil {
				if err := hook(off); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if m.options.PullWorkers > 0 {
		m.puller = chunks.NewPuller(
			m.ctx,
			syncedReadWriter,
			m.options.ChunkSize,
			chunkCount,
			func(offset int64) int64 {
				return 1
			},
		)

		if !m.options.PullFirst {
			m.wg.Add(1)
			go func() {
				defer m.wg.Done()

				if err := m.puller.Wait(); err != nil {
					m.errs <- err

					return
				}
			}()
		}

		if err := m.puller.Open(m.options.PullWorkers); err != nil {
			return "", 0, err
		}

		if m.options.PullFirst {
			if err := m.puller.Wait(); err != nil {
				return "", 0, err
			}
		}
	}

	arbitraryReadWriter := chunks.NewArbitraryReadWriterAt(syncedReadWriter, m.options.ChunkSize)

	m.syncer = bbackend.NewReaderAtBackend(
		arbitraryReadWriter,
		func() (int64, error) {
			return size, nil
		},
		func() error {
			if hook := m.hooks.OnBeforeSync; hook != nil {
				if err := hook(); err != nil {
					return err
				}
			}

			// We only ever touch the remote if we want to push
			if m.options.PushWorkers > 0 {
				_, err := local.(*chunks.Pusher).Flush()
				if err != nil {
					return err
				}

				if err := m.remote.Sync(); err != nil {
					return err
				}
			}

			if err := m.local.Sync(); err != nil {
				return err
			}

			return nil
		},
		m.options.Verbose,
	)

	m.dev = device.NewDevice(
		m.syncer,
		m.serverFile,

		m.serverOptions,
		m.clientOptions,
	)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		if err := m.dev.Wait(); err != nil {
			m.errs <- err

			return
		}
	}()

	if err := m.dev.Open(); err != nil {
		return "", 0, err
	}

	return devicePath, size, nil
}

func (m *PathFrontend) Close() error {
	if m.syncer != nil {
		_ = m.syncer.Sync()
	}

	if hook := m.hooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	if m.dev != nil {
		_ = m.dev.Close()
	}

	if hook := m.hooks.OnAfterClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	if m.puller != nil {
		_ = m.puller.Close()
	}

	if m.pusher != nil {
		_ = m.pusher.Close()
	}

	if m.serverFile != nil {
		_ = m.serverFile.Close()
	}

	m.wg.Wait()

	close(m.errs)

	return nil
}

func (m *PathFrontend) Sync() error {
	return m.syncer.Sync()
}
