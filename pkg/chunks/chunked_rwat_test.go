package chunks

import (
	"bytes"
	"os"
	"testing"
)

func TestChunkedReadWriterAt(t *testing.T) {
	tests := []struct {
		name             string
		chunkSize        int64
		chunks           int64
		input            []byte
		offset           int64
		expectedData     []byte
		expectedN        int
		expectedReadErr  error
		expectedWriteErr error
	}{
		{
			name:             "Write and read with chunks of size 4",
			chunkSize:        4,
			chunks:           1,
			input:            []byte("1234"),
			offset:           0,
			expectedData:     []byte("1234"),
			expectedN:        4,
			expectedReadErr:  nil,
			expectedWriteErr: nil,
		},
		{
			name:             "Write and read with chunks of size 1",
			chunkSize:        1,
			chunks:           4,
			input:            []byte("1"),
			offset:           0,
			expectedData:     []byte("1"),
			expectedN:        1,
			expectedReadErr:  nil,
			expectedWriteErr: nil,
		},
		{
			name:             "Write and read invalid input (too large)",
			chunkSize:        4,
			chunks:           1,
			input:            []byte("12345"),
			offset:           0,
			expectedData:     nil,
			expectedN:        0,
			expectedReadErr:  ErrInvalidOffset,
			expectedWriteErr: ErrInvalidOffset,
		},
		{
			name:             "Write and read invalid input (too small)",
			chunkSize:        4,
			chunks:           1,
			input:            []byte("123"),
			offset:           0,
			expectedData:     nil,
			expectedN:        0,
			expectedReadErr:  ErrInvalidOffset,
			expectedWriteErr: ErrInvalidOffset,
		},
		{
			name:             "Write and read at invalid offset (too large)",
			chunkSize:        4,
			chunks:           1,
			input:            []byte("1234"),
			offset:           6,
			expectedData:     nil,
			expectedN:        0,
			expectedReadErr:  ErrInvalidOffset,
			expectedWriteErr: ErrInvalidOffset,
		},
		{
			name:             "Write and read at invalid offset (too small)",
			chunkSize:        4,
			chunks:           1,
			input:            []byte("1234"),
			offset:           6,
			expectedData:     nil,
			expectedN:        0,
			expectedReadErr:  ErrInvalidOffset,
			expectedWriteErr: ErrInvalidOffset,
		},
		{
			name:             "Write and read the second chunk",
			chunkSize:        4,
			chunks:           3,
			input:            []byte("1234"),
			offset:           4,
			expectedData:     []byte("1234"),
			expectedN:        4,
			expectedReadErr:  nil,
			expectedWriteErr: nil,
		},
		{
			name:             "Write and read the last chunk",
			chunkSize:        2,
			chunks:           3,
			input:            []byte("56"),
			offset:           4,
			expectedData:     []byte("56"),
			expectedN:        2,
			expectedReadErr:  nil,
			expectedWriteErr: nil,
		},
		{
			name:             "Read beyond EOF",
			chunkSize:        4,
			chunks:           1,
			input:            nil,
			offset:           4,
			expectedData:     nil,
			expectedN:        0,
			expectedReadErr:  ErrInvalidReadSize,
			expectedWriteErr: nil,
		},
		{
			name:             "Write beyond EOF",
			chunkSize:        4,
			chunks:           2,
			input:            []byte("6789"),
			offset:           8,
			expectedData:     nil,
			expectedN:        0,
			expectedReadErr:  nil,
			expectedWriteErr: ErrInvalidWriteSize,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(f.Name())

			if err := f.Truncate(tc.chunkSize * tc.chunks); err != nil {
				t.Fatal(err)
			}

			crw := NewChunkedReadWriterAt(f, tc.chunkSize, tc.chunks)
			wbuf := make([]byte, tc.chunkSize)

			if tc.input != nil {
				wn, werr := crw.WriteAt(tc.input, tc.offset)
				if werr != tc.expectedWriteErr {
					t.Errorf("WriteAt error: expected %v, got %v", tc.expectedWriteErr, werr)
				}

				if wn != tc.expectedN {
					t.Errorf("WriteAt bytes written: expected %v, got %v", tc.expectedN, wn)
				}
			}

			if tc.expectedData != nil {
				rn, rerr := crw.ReadAt(wbuf, tc.offset)
				if rerr != tc.expectedReadErr {
					t.Errorf("ReadAt error: expected %v, got %v", tc.expectedReadErr, rerr)
				}

				if rn != tc.expectedN {
					t.Errorf("ReadAt bytes read: expected %v, got %v", tc.expectedN, rn)
				}

				if !bytes.Equal(wbuf, tc.expectedData) {
					t.Errorf("ReadAt data: expected %v, got %v", tc.expectedData, wbuf)
				}
			}
		})
	}
}
