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
	data      []byte
	chunkSize int64
}

func NewChunkedReadWriterAt(chunkSize, chunks int64) *ChunkedReadWriterAt {
	return &ChunkedReadWriterAt{
		make([]byte, chunkSize*chunks),
		chunkSize,
	}
}

func (c *ChunkedReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off%c.chunkSize != 0 || int64(len(p)) != c.chunkSize {
		return 0, ErrInvalidOffset
	}

	if off < 0 || off >= int64(len(c.data)) {
		return 0, ErrInvalidReadSize
	}

	n = copy(p, c.data[off:off+c.chunkSize])
	if n != len(p) {
		return n, io.EOF
	}

	return n, nil
}

func (c *ChunkedReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off%c.chunkSize != 0 || int64(len(p)) != c.chunkSize {
		return 0, ErrInvalidOffset
	}

	if off < 0 || off+c.chunkSize > int64(len(c.data)) {
		return 0, ErrInvalidWriteSize
	}

	n = copy(c.data[off:off+c.chunkSize], p)
	if n != len(p) {
		return n, io.ErrShortWrite
	}

	return n, nil
}
