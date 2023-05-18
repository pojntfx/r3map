package services

import (
	"context"
	"log"

	"github.com/pojntfx/go-nbd/pkg/backend"
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
}

func NewBackend(b backend.Backend, verbose bool) *Backend {
	return &Backend{b, verbose}
}

func (b *Backend) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", length, off)
	}

	r = ReadAtResponse{
		P: make([]byte, length),
	}

	r.N, err = b.b.ReadAt(r.P, off)

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
