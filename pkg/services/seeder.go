package services

import (
	"context"
	"errors"
	"log"

	"github.com/pojntfx/go-nbd/pkg/backend"
)

var (
	ErrMaxChunkSizeExceeded = errors.New("max chunk size exceeded")
)

type SeederRemote struct {
	ReadAt func(context context.Context, length int, off int64) (r ReadAtResponse, err error)
	Size   func(context context.Context) (int64, error)
	Track  func(context context.Context) error
	Sync   func(context context.Context) ([]int64, error)
	Close  func(context context.Context) error
}

type Seeder struct {
	b       backend.Backend
	verbose bool

	track func() error
	sync  func() ([]int64, error)
	close func() error

	maxChunkSize int64
}

func NewSeeder(
	b backend.Backend,
	verbose bool,
	track func() error,
	sync func() ([]int64, error),
	close func() error,
	maxChunkSize int64,
) *Seeder {
	if maxChunkSize <= 0 {
		maxChunkSize = MaxChunkSize
	}

	return &Seeder{b, verbose, track, sync, close, maxChunkSize}
}

func (b *Seeder) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
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

func (b *Seeder) Size(context context.Context) (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.b.Size()
}

func (b *Seeder) Track(context context.Context) error {
	if b.verbose {
		log.Println("Track()")
	}

	return b.track()
}

func (b *Seeder) Sync(context context.Context) ([]int64, error) {
	if b.verbose {
		log.Println("Sync()")
	}

	return b.sync()
}

func (b *Seeder) Close(context context.Context) error {
	if b.verbose {
		log.Println("Close()")
	}

	return b.close()
}
