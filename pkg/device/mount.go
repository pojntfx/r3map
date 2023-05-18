package device

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
	"github.com/pojntfx/r3map/pkg/utils"
)

type MountOptions struct {
	ChunkSize int64

	PullWorkers int64
	PullFirst   bool

	PushWorkers  int64
	PushInterval time.Duration

	Verbose bool
}

type MountHooks struct {
	OnBeforeSync func() error

	OnBeforeClose func() error
	OnAfterClose  func() error
}

type Mount struct {
	ctx context.Context

	remote,
	local,
	syncer backend.Backend

	mountOptions *MountOptions
	mountHooks   *MountHooks

	serverOptions *server.Options
	clientOptions *client.Options

	serverFile *os.File
	pusher     *chunks.Pusher
	puller     *chunks.Puller
	dev        *Device

	wg   sync.WaitGroup
	errs chan error
}

func NewMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	mountOptions *MountOptions,
	mountHooks *MountHooks,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *Mount {
	if mountOptions == nil {
		mountOptions = &MountOptions{}
	}

	if mountOptions.ChunkSize <= 0 {
		mountOptions.ChunkSize = 4096
	}

	if mountOptions.PushInterval == 0 {
		mountOptions.PushInterval = time.Second * 20
	}

	if mountHooks == nil {
		mountHooks = &MountHooks{}
	}

	return &Mount{
		ctx: ctx,

		remote: remote,
		local:  local,

		mountOptions: mountOptions,
		mountHooks:   mountHooks,

		serverOptions: serverOptions,
		clientOptions: clientOptions,

		errs: make(chan error),
	}
}

func (m *Mount) Wait() error {
	for err := range m.errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Mount) Open() (string, int64, error) {
	size, err := m.remote.Size()
	if err != nil {
		return "", 0, err
	}
	chunkCount := size / m.mountOptions.ChunkSize

	devicePath, err := utils.FindUnusedNBDDevice(time.Millisecond * 50)
	if err != nil {
		return "", 0, err
	}

	m.serverFile, err = os.Open(devicePath)
	if err != nil {
		return "", 0, err
	}

	var local chunks.ReadWriterAt
	if m.mountOptions.PushWorkers > 0 {
		m.pusher = chunks.NewPusher(
			m.ctx,
			chunks.NewChunkedReadWriterAt(m.local, m.mountOptions.ChunkSize, chunkCount),
			m.remote,
			m.mountOptions.ChunkSize,
			m.mountOptions.PushInterval,
		)

		m.wg.Add(1)
		go func() {
			defer m.wg.Done()

			if err := m.pusher.Wait(); err != nil {
				m.errs <- err

				return
			}
		}()

		if err := m.pusher.Open(m.mountOptions.PushWorkers); err != nil {
			return "", 0, err
		}

		local = m.pusher
	} else {
		local = chunks.NewChunkedReadWriterAt(m.local, m.mountOptions.ChunkSize, chunkCount)
	}

	syncedReadWriter := chunks.NewSyncedReadWriterAt(m.remote, local, func(off int64) error {
		if m.mountOptions.PushWorkers > 0 {
			if err := local.(*chunks.Pusher).MarkOffsetPushable(off); err != nil {
				return err
			}
		}

		return nil
	})

	if m.mountOptions.PullWorkers > 0 {
		m.puller = chunks.NewPuller(
			m.ctx,
			syncedReadWriter,
			m.mountOptions.ChunkSize,
			chunkCount,
			func(offset int64) int64 {
				return 1
			},
		)

		if !m.mountOptions.PullFirst {
			m.wg.Add(1)
			go func() {
				defer m.wg.Done()

				if err := m.puller.Wait(); err != nil {
					m.errs <- err

					return
				}
			}()
		}

		if err := m.puller.Open(m.mountOptions.PullWorkers); err != nil {
			return "", 0, err
		}

		if m.mountOptions.PullFirst {
			if err := m.puller.Wait(); err != nil {
				return "", 0, err
			}
		}
	}

	arbitraryReadWriter := chunks.NewArbitraryReadWriterAt(syncedReadWriter, m.mountOptions.ChunkSize)

	m.syncer = bbackend.NewReaderAtBackend(
		arbitraryReadWriter,
		func() (int64, error) {
			return size, nil
		},
		func() error {
			if hook := m.mountHooks.OnBeforeSync; hook != nil {
				if err := hook(); err != nil {
					return err
				}
			}

			// We only ever touch the remote if we want to push
			if m.mountOptions.PushWorkers > 0 {
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
		m.mountOptions.Verbose,
	)

	m.dev = NewDevice(
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

func (m *Mount) Close() error {
	if m.syncer != nil {
		_ = m.syncer.Sync()
	}

	if hook := m.mountHooks.OnBeforeClose; hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	if m.dev != nil {
		_ = m.dev.Close()
	}

	if hook := m.mountHooks.OnAfterClose; hook != nil {
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

func (m *Mount) Sync() error {
	return m.syncer.Sync()
}
