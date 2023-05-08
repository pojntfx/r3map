package backend

import (
	"context"
	"log"

	"github.com/pojntfx/r3map/pkg/services"
)

type RPCBackend struct {
	ctx     context.Context
	remote  services.BackendRemote
	verbose bool
}

func NewRPCBackend(
	ctx context.Context,
	remote services.BackendRemote,
	verbose bool,
) *RPCBackend {
	return &RPCBackend{ctx, remote, verbose}
}

func (b *RPCBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	r, err := b.remote.ReadAt(b.ctx, len(p), off)
	if err != nil {
		return -1, err
	}

	n = r.N
	copy(p, r.P)

	return
}

func (b *RPCBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	return b.remote.WriteAt(b.ctx, p, off)
}

func (b *RPCBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.remote.Size(b.ctx)
}

func (b *RPCBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	return b.remote.Sync(b.ctx)
}
