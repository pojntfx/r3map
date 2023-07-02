package services

import (
	"context"
	"log"

	"github.com/pojntfx/go-nbd/pkg/backend"
)

const (
	MaxChunkSize = 32 * 1024 * 1024 // 32MB; this is theoretically the maximum size a single NBD packet can be, but realistically this will always be <32kB (which is what Go's internal `io.Copy` uses as it's buffer size)
)

type BackendRemote struct {
	ReadAt  func(context context.Context, length int, off int64) (r ReadAtResponse, err error)
	WriteAt func(context context.Context, p []byte, off int64) (n int, err error)
	Size    func(context context.Context) (int64, error)
	Sync    func(context context.Context) error
}

type ReadAtResponse struct {
	N int
	P []byte
}

type Backend struct {
	b       backend.Backend
	verbose bool

	maxChunkSize int64
}

func NewBackend(b backend.Backend, verbose bool, maxChunkSize int64) *Backend {
	return &Backend{b, verbose, maxChunkSize}
}

func (b *Backend) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", length, off)
	}

	if int64(length) > b.maxChunkSize {
		return ReadAtResponse{}, ErrMaxChunkSizeExceeded
	}

	r = ReadAtResponse{
		P: make([]byte, length),
	}

	n, err := b.b.ReadAt(r.P, off)
	if err != nil {
		return ReadAtResponse{}, err
	}
	r.N = n

	return
}

func (b *Backend) WriteAt(context context.Context, p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	return b.b.WriteAt(p, off)
}

func (b *Backend) Size(context context.Context) (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.b.Size()
}

func (b *Backend) Sync(context context.Context) error {
	if b.verbose {
		log.Println("Sync()")
	}

	return b.b.Sync()
}
