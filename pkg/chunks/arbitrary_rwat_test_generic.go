package chunks

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

type ReadWriteConfiguration struct {
	ChunkSizes  []int64
	ChunkCount  int64
	BufferSizes []int64
}

func TestArbitraryReadWriterAtGeneric(
	t *testing.T,
	getArbitraryReadWriterAt func(chunkSize, chunkCount int64) (ReadWriterAt, func() error, error),
	configurations []ReadWriteConfiguration,
) {
	for _, cfg := range configurations {
		for _, chunkSize := range cfg.ChunkSizes {
			for _, bufferSize := range cfg.BufferSizes {
				for offset := int64(0); offset < cfg.ChunkCount*chunkSize; offset += bufferSize {
					t.Run(fmt.Sprintf("read_uninitialized_chunkSize_%d_chunkCount_%d_bufferSize_%d_offset_%d", chunkSize, cfg.ChunkCount, bufferSize, offset), func(t *testing.T) {
						arw, free, err := getArbitraryReadWriterAt(chunkSize, cfg.ChunkCount)
						if err != nil {
							t.Fatal(err)
						}
						defer free()

						expectedData := make([]byte, bufferSize)

						buf := make([]byte, bufferSize)
						n, err := arw.ReadAt(buf, offset)
						if err != nil {
							t.Fatal(err)
						}

						if n != len(buf) {
							t.Errorf("read %d bytes, expected %d", n, len(buf))
						}

						if !bytes.Equal(buf, expectedData) {
							t.Errorf("got %v, want %v", buf, expectedData)
						}
					})

					t.Run(fmt.Sprintf("read_write_chunkSize_%d_chunkCount_%d_bufferSize_%d_offset_%d", chunkSize, cfg.ChunkCount, bufferSize, offset), func(t *testing.T) {
						arw, free, err := getArbitraryReadWriterAt(chunkSize, cfg.ChunkCount)
						if err != nil {
							t.Fatal(err)
						}
						defer free()

						expectedData := make([]byte, bufferSize)
						_, err = rand.Read(expectedData)
						if err != nil {
							t.Fatal(err)
						}

						n, err := arw.WriteAt(expectedData, offset)
						if err != nil {
							t.Fatal(err)
						}

						if n != len(expectedData) {
							t.Errorf("wrote %d bytes, expected %d", n, len(expectedData))
						}

						buf := make([]byte, bufferSize)

						n, err = arw.ReadAt(buf, offset)
						if err != nil {
							t.Fatal(err)
						}

						if n != len(buf) {
							t.Errorf("read %d bytes, expected %d", n, len(buf))
						}

						if !bytes.Equal(buf, expectedData) {
							t.Errorf("got %v, want %v", buf, expectedData)
						}
					})
				}
			}
		}
	}
}
