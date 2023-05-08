package services

import (
	"context"

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
	b backend.Backend
}

func NewBackend(b backend.Backend) *Backend {
	return &Backend{b}
}

func (w *Backend) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
	r = ReadAtResponse{
		P: make([]byte, length),
	}

	r.N, err = w.b.ReadAt(r.P, off)

	return
}

func (b *Backend) WriteAt(context context.Context, p []byte, off int64) (n int, err error) {
	return b.b.WriteAt(p, off)
}

func (b *Backend) Size(context context.Context) (int64, error) {
	return b.b.Size()
}

func (b *Backend) Sync(context context.Context) error {
	return b.b.Sync()
}
