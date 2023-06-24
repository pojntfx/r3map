package chunks

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
)

func TestPuller(t *testing.T) {
	tests := []struct {
		name                   string
		chunkSize              int64
		chunks                 int64
		workers                int64
		data                   [][]byte
		pullPriority           func(off int64) int64
		dirtyOffsets           []int64
		waitTillFullyAvailable bool
		modifyData             bool
	}{
		{
			name:      "Pull 1 chunk with 1 worker and generic pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   1,
			data:      [][]byte{[]byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 1 chunk with 2 workers and generic pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   2,
			data:      [][]byte{[]byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 2 chunks with 1 worker and generic pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   1,
			data:      [][]byte{[]byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 2 chunks with 2 workers and generic pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 1 chunk with 1 worker and linear pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   1,
			data:      [][]byte{[]byte("test")},
			pullPriority: func(off int64) int64 {
				return off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 1 chunk with 2 workers and linear pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   2,
			data:      [][]byte{[]byte("test")},
			pullPriority: func(off int64) int64 {
				return off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 2 chunks with 1 worker and linear pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   1,
			data:      [][]byte{[]byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 2 chunks with 2 workers and linear pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 1 chunk with 1 worker and decreasing pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   1,
			data:      [][]byte{[]byte("test")},
			pullPriority: func(off int64) int64 {
				return -off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 1 chunk with 2 workers and decreasing pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   2,
			data:      [][]byte{[]byte("test")},
			pullPriority: func(off int64) int64 {
				return -off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 2 chunks with 1 worker and decreasing pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   1,
			data:      [][]byte{[]byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return -off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull 2 chunks with 2 workers and decreasing pull priority heuristic",
			chunkSize: 4,
			chunks:    2,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return -off
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull with no chunks finalization",
			chunkSize: 4,
			chunks:    3,
			workers:   1,
			data:      [][]byte{[]byte("test"), []byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{},
		},
		{
			name:      "Pull with some chunks finalization",
			chunkSize: 4,
			chunks:    3,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{8},
		},
		{
			name:      "Pull with all chunks finalization",
			chunkSize: 4,
			chunks:    3,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets: []int64{0, 4, 8},
		},
		{
			name:      "Pull with all chunks finalization and full availability",
			chunkSize: 4,
			chunks:    3,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets:           []int64{0, 4, 8},
			waitTillFullyAvailable: true,
		},
		{
			name:      "Pull with dirty offsets and subsequent data modification",
			chunkSize: 4,
			chunks:    3,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets:           []int64{4},
			waitTillFullyAvailable: false,
			modifyData:             true,
		},
		{
			name:      "Pull with dirty offsets and subsequent data modification and full availability",
			chunkSize: 4,
			chunks:    3,
			workers:   2,
			data:      [][]byte{[]byte("test"), []byte("test"), []byte("test")},
			pullPriority: func(off int64) int64 {
				return 1
			},
			dirtyOffsets:           []int64{4},
			waitTillFullyAvailable: true,
			modifyData:             true,
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

			for i, chunk := range tc.data {
				if _, werr := remote.WriteAt(chunk, int64(i)*tc.chunkSize); werr != nil {
					t.Fatal(err)
				}
			}

			local := NewChunkedReadWriterAt(localFile, tc.chunkSize, tc.chunks)

			var wg sync.WaitGroup
			if tc.waitTillFullyAvailable {
				wg.Add(1)
			}

			srw := NewSyncedReadWriterAt(remote, local, func(off int64) error {
				if (off / tc.chunkSize) >= tc.chunks-1 {
					if tc.waitTillFullyAvailable {
						wg.Done()
					}
				}

				return nil
			})

			ctx := context.Background()

			puller := NewPuller(
				ctx,
				srw,
				tc.chunkSize,
				tc.chunks,
				tc.pullPriority,
			)
			err = puller.Open(tc.workers)
			if err != nil {
				t.Fatal(err)
			}

			if tc.modifyData {
				for _, offset := range tc.dirtyOffsets {
					if _, werr := remote.WriteAt([]byte("modi"), offset); werr != nil {
						t.Fatal(err)
					}
				}
			}

			puller.Finalize(tc.dirtyOffsets)

			wg.Add(1)
			go func() {
				defer wg.Done()

				if err := puller.Wait(); err != nil {
					panic(err)
				}
			}()

			wg.Wait()

			if err := puller.Close(); err != nil {
				t.Fatal(err)
			}

			for i := int64(0); i < tc.chunks; i++ {
				var expectedData []byte
				isDirty := false
				for _, offset := range tc.dirtyOffsets {
					if i*tc.chunkSize == offset {
						isDirty = true
						break
					}
				}

				if tc.modifyData && isDirty {
					// If the chunk was a dirty chunk and was modified, expect "modi"
					expectedData = []byte("modi")
				} else if i < int64(len(tc.data)) {
					// If it's a regular chunk, expect the original data
					expectedData = tc.data[i]
				} else {
					// If it's an extra chunk (more chunks than data provided), expect zeros
					expectedData = make([]byte, tc.chunkSize)
				}

				localData := make([]byte, tc.chunkSize)
				if _, err := local.ReadAt(localData, i*tc.chunkSize); err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(localData, expectedData) {
					t.Errorf("Data pulled did not match expected. got %v, want %v", localData, expectedData)
				}
			}
		})
	}
}
