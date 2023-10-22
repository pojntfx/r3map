package mount

import (
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/go-nbd/pkg/server"
)

type DirectSliceMount struct {
	path *DirectPathMount

	devicePath string
	deviceFile *os.File

	slice     mmap.MMap
	mmapMount sync.Mutex

	b backend.Backend

	closeLock sync.Mutex
}

func NewDirectSliceMount(
	b backend.Backend,
	f *os.File,

	serverOptions *server.Options,
	clientOptions *client.Options,
) *DirectSliceMount {
	return &DirectSliceMount{
		path: NewDirectPathMount(
			b,
			f,

			serverOptions,
			clientOptions,
		),

		devicePath: f.Name(),

		b: b,
	}
}

func (d *DirectSliceMount) Wait() error {
	return d.path.Wait()
}

func (d *DirectSliceMount) Open() ([]byte, error) {
	size, err := d.b.Size()
	if err != nil {
		return []byte{}, err
	}

	if err := d.path.Open(); err != nil {
		return []byte{}, err
	}

	deviceFile, err := os.OpenFile(d.devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return []byte{}, err
	}
	d.deviceFile = deviceFile

	d.slice, err = mmap.MapRegion(d.deviceFile, int(size), mmap.RDWR, 0, 0)
	if err != nil {
		return []byte{}, err
	}

	// We _MUST_ lock this slice so that it does not get paged out
	// If it does, the Go GC tries to manage it, deadlocking _the entire runtime_
	if err := d.slice.Lock(); err != nil {
		return []byte{}, err
	}

	return d.slice, nil
}

func (d *DirectSliceMount) Close() error {
	d.closeLock.Lock()
	defer d.closeLock.Unlock()

	d.mmapMount.Lock()
	if d.slice != nil {
		_ = d.slice.Unlock()

		_ = d.slice.Unmap()

		d.slice = nil
	}
	d.mmapMount.Unlock()

	if d.deviceFile != nil {
		_ = d.deviceFile.Close()
	}

	return d.path.Close()
}

func (d *DirectSliceMount) Sync() error {
	d.mmapMount.Lock()
	if d.slice != nil {
		if err := d.slice.Flush(); err != nil {
			return err
		}
	}
	d.mmapMount.Unlock()

	return d.path.Sync()
}

func (d *DirectSliceMount) SwapBackend(b backend.Backend) {
	d.path.e.Backend = b
}
