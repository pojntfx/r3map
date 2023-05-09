package chunks

import (
	"bytes"
	"testing"
)

func TestArbitraryReadWriterAt(t *testing.T) {
	tests := []struct {
		name         string
		chunkSize    int64
		chunks       int64
		input        []byte
		offset       int64
		rwBufferSize int
		expectedData []byte
		readErr      error
		writeErr     error
	}{
		{
			name:         "Write and read at offset 0, entire chunk",
			chunkSize:    4,
			chunks:       3,
			input:        []byte("1234"),
			offset:       0,
			rwBufferSize: 4,
			expectedData: []byte("1234"),
			readErr:      nil,
			writeErr:     nil,
		},
		{
			name:         "Write and read at offset 0, partial chunk",
			chunkSize:    4,
			chunks:       3,
			input:        []byte("12"),
			offset:       0,
			rwBufferSize: 2,
			expectedData: []byte("12"),
			readErr:      nil,
			writeErr:     nil,
		},
		{
			name:         "Write and read across two chunks",
			chunkSize:    4,
			chunks:       3,
			input:        []byte("3456"),
			offset:       2,
			rwBufferSize: 4,
			expectedData: []byte("3456"),
			readErr:      nil,
			writeErr:     nil,
		},
		{
			name:         "Write and read big buffer spanning multiple chunks",
			chunkSize:    4,
			chunks:       3,
			input:        []byte("123456789ABC"),
			offset:       0,
			rwBufferSize: 12,
			expectedData: []byte("123456789ABC"),
			readErr:      nil,
			writeErr:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			crw := NewChunkedReadWriterAt(tc.chunkSize, tc.chunks)
			arw := NewArbitraryReadWriterAt(crw, tc.chunkSize)

			if tc.input != nil {
				wn, werr := arw.WriteAt(tc.input, tc.offset)
				if werr != tc.writeErr {
					t.Errorf("WriteAt error: expected %v, got %v", tc.writeErr, werr)
				}

				if wn != len(tc.input) {
					t.Errorf("WriteAt bytes written: expected %v, got %v", len(tc.input), wn)
				}
			}

			if tc.expectedData != nil || tc.readErr != nil {
				rbuf := make([]byte, tc.rwBufferSize)
				rn, rerr := arw.ReadAt(rbuf, tc.offset)
				if rerr != tc.readErr {
					t.Errorf("ReadAt error: expected %v, got %v", tc.readErr, rerr)
				}

				if rn != len(tc.expectedData) {
					t.Errorf("ReadAt bytes read: expected %v, got %v", len(tc.expectedData), rn)
				}

				if tc.expectedData != nil {
					rbuf = rbuf[:rn]
					if !bytes.Equal(rbuf, tc.expectedData) {
						t.Errorf("ReadAt data: expected %v, got %v", tc.expectedData, rbuf)
					}
				}
			}
		})
	}
}
