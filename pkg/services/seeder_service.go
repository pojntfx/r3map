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

type SeederService struct {
	b       backend.Backend
	verbose bool

	track func() error
	sync  func() ([]int64, error)
	close func() error

	maxChunkSize int64
}

func NewSeederService(
	b backend.Backend,
	verbose bool,
	track func() error,
	sync func() ([]int64, error),
	close func() error,
	maxChunkSize int64,
) *SeederService {
	if maxChunkSize <= 0 {
		maxChunkSize = MaxChunkSize
	}

	return &SeederService{b, verbose, track, sync, close, maxChunkSize}
}

func (b *SeederService) ReadAt(context context.Context, length int, off int64) (r ReadAtResponse, err error) {
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

func (b *SeederService) Track(context context.Context) error {
	if b.verbose {
		log.Println("Track()")
	}

	return b.track()
}

func (b *SeederService) Sync(context context.Context) ([]int64, error) {
	if b.verbose {
		log.Println("Sync()")
	}

	return b.sync()
}

func (b *SeederService) Close(context context.Context) error {
	if b.verbose {
		log.Println("Close()")
	}

	return b.close()
}
