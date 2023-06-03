package chunks

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func TestChunkedReadWriterAtGeneric(
	t *testing.T,
	getReadWriterAt func(chunkSize, chunkCount int64) (ReadWriterAt, func() error, error),
	validChunkSizes,
	validChunkCounts []int64,
) {
	for _, chunkSize := range validChunkSizes {
		for _, chunkCount := range validChunkCounts {
			for offset := int64(0); offset < chunkSize*chunkCount; offset += chunkSize {
				writeData := make([]byte, chunkSize)
				_, err := rand.Read(writeData)
				if err != nil {
					t.Fatal(err)
				}

				for _, tc := range []struct {
					name      string
					offset    int64
					readSize  int64
					writeData []byte
					wantBytes []byte
				}{
					{
						name:      fmt.Sprintf("read_uninitialized_chunk_offset_%d", offset),
						offset:    int64(offset),
						readSize:  chunkSize,
						wantBytes: make([]byte, chunkSize),
					},
					{
						name:      fmt.Sprintf("read_write_chunk_offset_%d", offset),
						offset:    int64(offset),
						readSize:  chunkSize,
						writeData: writeData,
						wantBytes: writeData,
					},
				} {
					t.Run(fmt.Sprintf("%s_chunkSize_%d_chunks_%d", tc.name, chunkSize, chunkCount), func(t *testing.T) {
						rwa, free, err := getReadWriterAt(chunkSize, chunkCount)
						if err != nil {
							t.Fatal(err)
						}
						defer free()

						if len(tc.writeData) > 0 {
							n, err := rwa.WriteAt(tc.writeData, tc.offset)
							if err != nil {
								t.Fatal(err)
							}

							if n != len(tc.writeData) {
								t.Errorf("wrote %d bytes, expected %d", n, len(tc.writeData))
							}
						}

						buf := make([]byte, tc.readSize)

						n, err := rwa.ReadAt(buf, tc.offset)
						if err != nil {
							t.Fatal(err)
						}

						if n != len(buf) {
							t.Errorf("read %d bytes, expected %d", n, len(buf))
						}
						if !bytes.Equal(buf, tc.wantBytes) {
							t.Errorf("got %v, want %v", buf, tc.wantBytes)
						}
					})
				}
			}
		}
	}
}
