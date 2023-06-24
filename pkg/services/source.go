package services

import (
	"context"
	"log"

	"github.com/pojntfx/go-nbd/pkg/backend"
)

type SeederRemote struct {
	ReadAt func(context context.Context, length int, off int64) (r ReadAtResponse, err error)
	Size   func(context context.Context) (int64, error)
	Track  func(context context.Context) error
	Flush  func(context context.Context) ([]int64, error)
	Close  func(context context.Context) error
}

type Seeder struct {
	b       backend.Backend
	verbose bool

	track func() error
	flush func() ([]int64, error)
	close func() error
}

func NewSeeder(
	b backend.Backend,
	verbose bool,
	track func() error,
	flush func() ([]int64, error),
	close func() error,
) *Seeder {
	return &Seeder{b, verbose, track, flush, close}
}

func (b *Seeder) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", length, off)
	}

	r = ReadAtResponse{
		P: make([]byte, length),
	}

	r.N, err = b.b.ReadAt(r.P, off)

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

func (b *Seeder) Flush(context context.Context) ([]int64, error) {
	if b.verbose {
		log.Println("Flush()")
	}

	return b.flush()
}

func (b *Seeder) Close(context context.Context) error {
	if b.verbose {
		log.Println("Close()")
	}

	return b.close()
}
