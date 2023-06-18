package services

import (
	"context"
	"log"

	"github.com/pojntfx/go-nbd/pkg/backend"
)

type SourceRemote struct {
	ReadAt func(context context.Context, length int, off int64) (r ReadAtResponse, err error)
	Size   func(context context.Context) (int64, error)
	Flush  func(context context.Context) ([]int64, error)
	Track  func(context context.Context) error
}

type Source struct {
	b       backend.Backend
	verbose bool

	flush func() ([]int64, error)
	track func() error
}

func NewSource(
	b backend.Backend,
	verbose bool,
	flush func() ([]int64, error),
	track func() error,
) *Source {
	return &Source{b, verbose, flush, track}
}

func (b *Source) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", length, off)
	}

	r = ReadAtResponse{
		P: make([]byte, length),
	}

	r.N, err = b.b.ReadAt(r.P, off)

	return
}

func (b *Source) Size(context context.Context) (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.b.Size()
}

func (b *Source) Flush(context context.Context) ([]int64, error) {
	if b.verbose {
		log.Println("Flush()")
	}

	return b.flush()
}

func (b *Source) Track(context context.Context) error {
	if b.verbose {
		log.Println("Track()")
	}

	return b.track()
}
