package chunks

import (
	"bytes"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestTrackingReadWriterAt(t *testing.T) {
	tests := []struct {
		name            string
		inputs          [][]byte
		offsets         []int64
		rwBufferSize    int
		expectedData    [][][]byte
		expectedOffsets [][]int64
		preTrackWrites  int
		syncEachWrite   []int
	}{
		{
			name:            "Write and track one chunk",
			inputs:          [][]byte{[]byte("1234")},
			offsets:         []int64{0},
			rwBufferSize:    4,
			expectedData:    [][][]byte{{[]byte("1234")}},
			expectedOffsets: [][]int64{{0}},
			syncEachWrite:   []int{1},
		},
		{
			name:            "Write and track two chunks",
			inputs:          [][]byte{[]byte("1234"), []byte("5678")},
			offsets:         []int64{0, 4},
			rwBufferSize:    4,
			expectedData:    [][][]byte{{[]byte("1234"), []byte("5678")}},
			expectedOffsets: [][]int64{{0, 4}},
			syncEachWrite:   []int{2},
		},
		{
			name:            "Write and track three chunks",
			inputs:          [][]byte{[]byte("1234"), []byte("5678"), []byte("9012")},
			offsets:         []int64{0, 4, 8},
			rwBufferSize:    4,
			expectedData:    [][][]byte{{[]byte("1234"), []byte("5678"), []byte("9012")}},
			expectedOffsets: [][]int64{{0, 4, 8}},
			syncEachWrite:   []int{3},
		},
		{
			name:            "Write to the same offset twice",
			inputs:          [][]byte{[]byte("1234"), []byte("5678")},
			offsets:         []int64{0, 0},
			rwBufferSize:    4,
			expectedData:    [][][]byte{{[]byte("5678")}},
			expectedOffsets: [][]int64{{0}},
			syncEachWrite:   []int{2},
		},
		{
			name:            "Track only after calling Track",
			inputs:          [][]byte{[]byte("1234"), []byte("5678")},
			offsets:         []int64{0, 4},
			rwBufferSize:    4,
			expectedData:    [][][]byte{{[]byte("5678")}},
			expectedOffsets: [][]int64{{4}},
			preTrackWrites:  1,
		},
		{
			name:            "Writing and syncing twice only returns the second delta",
			inputs:          [][]byte{[]byte("1234"), []byte("5678"), []byte("9012")},
			offsets:         []int64{0, 4, 8},
			rwBufferSize:    4,
			expectedData:    [][][]byte{{[]byte("1234"), []byte("5678")}, {[]byte("9012")}},
			expectedOffsets: [][]int64{{0, 4}, {8}},
			syncEachWrite:   []int{2, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(f.Name())

			trw := NewTrackingReadWriterAt(f)

			if tc.preTrackWrites > 0 {
				for i := 0; i < tc.preTrackWrites; i++ {
					_, err := trw.WriteAt(tc.inputs[i], tc.offsets[i])
					if err != nil {
						t.Errorf("Pre-Track WriteAt error: got %v", err)
					}
				}
			}

			start := tc.preTrackWrites
			for i, syncEachWrite := range tc.syncEachWrite {
				trw.Track()

				for j := start; j < start+syncEachWrite; j++ {
					wn, err := trw.WriteAt(tc.inputs[j], tc.offsets[j])
					if err != nil {
						t.Errorf("WriteAt error: got %v", err)
					}

					if wn != len(tc.inputs[j]) {
						t.Errorf("WriteAt bytes written: expected %v, got %v", len(tc.inputs[j]), wn)
					}
				}

				dirtyOffsets := trw.Sync()

				sort.Slice(tc.expectedOffsets[i], func(x, y int) bool { return tc.expectedOffsets[i][x] < tc.expectedOffsets[i][y] })
				sort.Slice(dirtyOffsets, func(x, y int) bool { return dirtyOffsets[x] < dirtyOffsets[y] })

				if !reflect.DeepEqual(tc.expectedOffsets[i], dirtyOffsets) {
					t.Errorf("Sync offsets: expected %v, got %v", tc.expectedOffsets[i], dirtyOffsets)
				}

				start += syncEachWrite
			}

			rbuf := make([]byte, tc.rwBufferSize)
			start = 0
			for i, syncEachWrite := range tc.syncEachWrite {
				for j := start; j < start+syncEachWrite; j++ {
					rn, err := trw.ReadAt(rbuf, tc.offsets[j])
					if err != nil {
						t.Errorf("ReadAt error: got %v", err)
					}

					if j-start < len(tc.expectedData[i]) {
						expected := tc.expectedData[i][j-start]
						if rn != len(expected) {
							t.Errorf("ReadAt bytes read: expected %v, got %v", len(expected), rn)
						}

						rbuf = rbuf[:rn]
						if !bytes.Equal(rbuf, expected) {
							t.Errorf("ReadAt data: expected %v, got %v", expected, rbuf)
						}
					}
				}

				start += syncEachWrite
			}
		})
	}
}
