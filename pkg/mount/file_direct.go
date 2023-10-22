package mount

import (
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type DirectFileMount struct {
	path *DirectPathMount

	devicePath string
	deviceFile *os.File

	closeLock sync.Mutex
}

func NewDirectFileMount(
	b backend.Backend,
	f *os.File,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *DirectFileMount {
	return &DirectFileMount{
		path: NewDirectPathMount(
			b,
			f,

			serverOptions,
			clientOptions,
		),

		devicePath: f.Name(),
	}
}

func (d *DirectFileMount) Wait() error {
	return d.path.Wait()
}

func (d *DirectFileMount) Open() (*os.File, error) {
	if err := d.path.Open(); err != nil {
		return nil, err
	}

	deviceFile, err := os.OpenFile(d.devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	d.deviceFile = deviceFile

	return d.deviceFile, nil
}

func (d *DirectFileMount) Close() error {
	d.closeLock.Lock()
	defer d.closeLock.Unlock()

	if d.deviceFile != nil {
		_ = d.deviceFile.Close()

		d.deviceFile = nil
	}

	return d.path.Close()
}

func (d *DirectFileMount) Sync() error {
	if d.deviceFile != nil {
		if err := d.deviceFile.Sync(); err != nil {
			return err
		}
	}

	return d.path.Sync()
}

func (d *DirectFileMount) SwapBackend(b backend.Backend) {
	d.path.e.Backend = b
}
