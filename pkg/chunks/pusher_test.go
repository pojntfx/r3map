package chunks

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

func TestPusher(t *testing.T) {
	tests := []struct {
		name                string
		chunkSize           int64
		chunks              int64
		pushPeriod          time.Duration
		data                [][]byte
		markOffsetsPushable bool
	}{
		{
			name:                "Push 1 chunk",
			chunkSize:           4,
			chunks:              2,
			pushPeriod:          time.Second,
			data:                [][]byte{[]byte("test")},
			markOffsetsPushable: true,
		},
		{
			name:                "Push 2 chunks",
			chunkSize:           4,
			chunks:              2,
			pushPeriod:          time.Second,
			data:                [][]byte{[]byte("test"), []byte("test")},
			markOffsetsPushable: true,
		},
		{
			name:                "Do not push chunks without MarkOffsetPushable",
			chunkSize:           4,
			chunks:              2,
			pushPeriod:          time.Second,
			data:                [][]byte{[]byte("test"), []byte("test")},
			markOffsetsPushable: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			remoteFile, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(remoteFile.Name())

			if err := remoteFile.Truncate(tc.chunkSize * tc.chunks); err != nil {
				t.Fatal(err)
			}

			localFile, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(localFile.Name())

			if err := localFile.Truncate(tc.chunkSize * tc.chunks); err != nil {
				t.Fatal(err)
			}

			remote := NewChunkedReadWriterAt(remoteFile, tc.chunkSize, tc.chunks)
			local := NewChunkedReadWriterAt(localFile, tc.chunkSize, tc.chunks)

			ctx := context.Background()
			pusher := NewPusher(
				ctx,
				local,
				remote,
				tc.chunkSize,
				tc.pushPeriod,
			)
			err = pusher.Init()
			if err != nil {
				t.Fatal(err)
			}

			if tc.markOffsetsPushable {
				for i := int64(0); i < tc.chunks; i++ {
					if err := pusher.MarkOffsetPushable(i * tc.chunkSize); err != nil {
						t.Fatal(err)
					}
				}
			}

			go func() {
				if err := pusher.Wait(); err != nil {
					panic(err)
				}
			}()

			for i, chunk := range tc.data {
				if _, werr := pusher.WriteAt(chunk, int64(i)*tc.chunkSize); werr != nil {
					t.Fatal(err)
				}
			}

			if err := pusher.Close(); err != nil {
				t.Fatal(err)
			}

			if !tc.markOffsetsPushable {
				expectedRemoteData := make([]byte, tc.chunkSize)
				for i, _ := range tc.data {
					remoteData := make([]byte, len(expectedRemoteData))
					if _, err := remote.ReadAt(remoteData, int64(i)*tc.chunkSize); err != nil {
						t.Fatal(err)
					}

					if !bytes.Equal(remoteData, expectedRemoteData) {
						t.Errorf("Chunks were pushed without calling MarkOffsetPushable. got %v, want %v", remoteData, expectedRemoteData)
					}
				}
			} else {
				for i, chunk := range tc.data {
					remoteData := make([]byte, len(chunk))
					if _, err := remote.ReadAt(remoteData, int64(i)*tc.chunkSize); err != nil {
						t.Fatal(err)
					}

					if !bytes.Equal(remoteData, chunk) {
						t.Errorf("Data pushed did not match expected. got %v, want %v", remoteData, tc.data)
					}
				}
			}
		})
	}
}
