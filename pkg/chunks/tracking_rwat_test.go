package chunks

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestTrackingReadWriterAt(t *testing.T) {
	tests := []struct {
		name            string
		inputs          [][]byte
		offsets         []int64
		rwBufferSize    int
		expectedData    [][]byte
		expectedOffsets []int64
		readErr         error
		writeErr        error
	}{
		{
			name:            "Write and track one chunk",
			inputs:          [][]byte{[]byte("1234")},
			offsets:         []int64{0},
			rwBufferSize:    4,
			expectedData:    [][]byte{[]byte("1234")},
			expectedOffsets: []int64{0},
			readErr:         nil,
			writeErr:        nil,
		},
		{
			name:            "Write and track two chunks",
			inputs:          [][]byte{[]byte("1234"), []byte("5678")},
			offsets:         []int64{0, 4},
			rwBufferSize:    4,
			expectedData:    [][]byte{[]byte("1234"), []byte("5678")},
			expectedOffsets: []int64{0, 4},
			readErr:         nil,
			writeErr:        nil,
		},
		{
			name:            "Write and track three chunks",
			inputs:          [][]byte{[]byte("1234"), []byte("5678"), []byte("9012")},
			offsets:         []int64{0, 4, 8},
			rwBufferSize:    4,
			expectedData:    [][]byte{[]byte("1234"), []byte("5678"), []byte("9012")},
			expectedOffsets: []int64{0, 4, 8},
			readErr:         nil,
			writeErr:        nil,
		},
		{
			name:            "Write and track same offset twice",
			inputs:          [][]byte{[]byte("1234"), []byte("5678")},
			offsets:         []int64{0, 0},
			rwBufferSize:    4,
			expectedData:    [][]byte{[]byte("5678")},
			expectedOffsets: []int64{0},
			readErr:         nil,
			writeErr:        nil,
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

			if len(tc.inputs) > 0 {
				trw.Track()

				for i, input := range tc.inputs {
					wn, werr := trw.WriteAt(input, tc.offsets[i])
					if werr != tc.writeErr {
						t.Errorf("WriteAt error: expected %v, got %v", tc.writeErr, werr)
					}

					if wn != len(input) {
						t.Errorf("WriteAt bytes written: expected %v, got %v", len(input), wn)
					}
				}

				dirtyOffsets := trw.Flush()
				if !reflect.DeepEqual(tc.expectedOffsets, dirtyOffsets) {
					t.Errorf("Flush offsets: expected %v, got %v", tc.expectedOffsets, dirtyOffsets)
				}
			}

			if len(tc.expectedData) > 0 || tc.readErr != nil {
				rbuf := make([]byte, tc.rwBufferSize)
				for i, expected := range tc.expectedData {
					rn, rerr := trw.ReadAt(rbuf, tc.offsets[i])
					if rerr != tc.readErr {
						t.Errorf("ReadAt error: expected %v, got %v", tc.readErr, rerr)
					}

					if rn != len(expected) {
						t.Errorf("ReadAt bytes read: expected %v, got %v", len(expected), rn)
					}

					rbuf = rbuf[:rn]
					if !bytes.Equal(rbuf, expected) {
						t.Errorf("ReadAt data: expected %v, got %v", expected, rbuf)
					}
				}
			}
		})
	}
}
