package device

import (
	"os"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type FileDevice struct {
	path *PathDevice

	devicePath string
	deviceFile *os.File
}

func NewFileDevice(
	b backend.Backend,
	f *os.File,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *FileDevice {
	return &FileDevice{
		path: &PathDevice{
			b: b,
			f: f,

			serverOptions: serverOptions,
			clientOptions: clientOptions,

			errs: make(chan error),
		},

		devicePath: f.Name(),
	}
}

func (d *FileDevice) Wait() error {
	return d.path.Wait()
}

func (d *FileDevice) Open() (*os.File, error) {
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

func (d *FileDevice) Close() error {
	if d.deviceFile != nil {
		_ = d.deviceFile.Close()

		d.deviceFile = nil
	}

	return d.path.Close()
}

func (d *FileDevice) Sync() error {
	if d.deviceFile != nil {
		if err := d.deviceFile.Sync(); err != nil {
			return err
		}
	}

	return d.path.Sync()
}
