package device

import (
	"context"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

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

type Mount struct {
	ctx context.Context

	remote,
	local,
	syncer backend.Backend

	mountOptions *MountOptions

	serverOptions *server.Options
	clientOptions *client.Options

	serverFile *os.File
	pusher     *chunks.Pusher
	puller     *chunks.Puller
	dev        *Device
	clientFile *os.File

	mount     []byte
	mmapMount sync.Mutex

	wg   sync.WaitGroup
	errs chan error
}

func NewMount(
	ctx context.Context,

	remote backend.Backend,
	local backend.Backend,

	mountOptions *MountOptions,

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

	return &Mount{
		ctx: ctx,

		remote: remote,
		local:  local,

		mountOptions: mountOptions,

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

func (m *Mount) Open() ([]byte, error) {
	size, err := m.remote.Size()
	if err != nil {
		return []byte{}, err
	}
	chunkCount := size / m.mountOptions.ChunkSize

	devicePath, err := utils.FindUnusedNBDDevice(time.Millisecond * 50)
	if err != nil {
		return []byte{}, err
	}

	m.serverFile, err = os.Open(devicePath)
	if err != nil {
		return []byte{}, err
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
			return []byte{}, err
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
			return []byte{}, err
		}

		if m.mountOptions.PullFirst {
			if err := m.puller.Wait(); err != nil {
				return []byte{}, err
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
			m.mmapMount.Lock()
			if m.mount != nil {
				if _, _, err := syscall.Syscall(
					syscall.SYS_MSYNC,
					uintptr(unsafe.Pointer(&m.mount[0])),
					uintptr(len(m.mount)),
					uintptr(syscall.MS_SYNC),
				); err != 0 {
					m.mmapMount.Unlock()

					return err
				}
			}
			m.mmapMount.Unlock()

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
		return []byte{}, err
	}

	m.clientFile, err = os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return []byte{}, err
	}

	m.mount, err = syscall.Mmap(
		int(m.clientFile.Fd()),
		0,
		int(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return []byte{}, err
	}

	return m.mount, nil
}

func (m *Mount) Close() error {
	if m.syncer != nil {
		_ = m.syncer.Sync()
	}

	if m.clientFile != nil {
		_ = m.clientFile.Close()
	}

	if m.dev != nil {
		_ = m.dev.Close()
	}

	m.mmapMount.Lock()
	if m.mount != nil {
		_ = syscall.Munmap(m.mount)

		m.mount = nil
	}
	m.mmapMount.Unlock()

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
