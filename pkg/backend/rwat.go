package backend

import (
	"log"

	"github.com/pojntfx/r3map/pkg/chunks"
)

type ReaderAtBackend struct {
	backend chunks.ReadWriterAt

	size func() (int64, error)
	sync func() error

	verbose bool
}

func NewReaderAtBackend(
	backend chunks.ReadWriterAt,

	size func() (int64, error),
	sync func() error,

	verbose bool,
) *ReaderAtBackend {
	return &ReaderAtBackend{
		backend,
		size,
		sync,
		verbose,
	}
}

func (b *ReaderAtBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	return b.backend.ReadAt(p, off)
}

func (b *ReaderAtBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	return b.backend.WriteAt(p, off)
}

func (b *ReaderAtBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.size()
}

func (b *ReaderAtBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	return b.sync()
}
