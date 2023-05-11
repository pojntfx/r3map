package chunks

import (
	"errors"
	"io"
)

var (
	ErrInvalidOffset    = errors.New("invalid offset")
	ErrInvalidReadSize  = errors.New("invalid read size")
	ErrInvalidWriteSize = errors.New("invalid write size")
)

type ChunkedReadWriterAt struct {
	backend   ReadWriterAt
	chunkSize int64
	chunks    int64
}

func NewChunkedReadWriterAt(backend ReadWriterAt, chunkSize, chunks int64) *ChunkedReadWriterAt {
	return &ChunkedReadWriterAt{
		backend,
		chunkSize,
		chunks,
	}
}

func (c *ChunkedReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off%c.chunkSize != 0 || int64(len(p)) != c.chunkSize {
		return 0, ErrInvalidOffset
	}

	if off < 0 || off >= int64(c.chunkSize*c.chunks) {
		return 0, ErrInvalidReadSize
	}

	n, err = c.backend.ReadAt(p, off)
	if err != nil {
		return 0, err
	}

	if n != len(p) {
		return n, io.EOF
	}

	return n, nil
}

func (c *ChunkedReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off%c.chunkSize != 0 || int64(len(p)) != c.chunkSize {
		return 0, ErrInvalidOffset
	}

	if off < 0 || off+c.chunkSize > int64(c.chunkSize*c.chunks) {
		return 0, ErrInvalidWriteSize
	}

	n, err = c.backend.WriteAt(p, off)
	if err != nil {
		return 0, err
	}

	if n != len(p) {
		return n, io.ErrShortWrite
	}

	return n, nil
}
