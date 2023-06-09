package chunks

import (
	"bytes"
	"os"
	"testing"
)

func TestSyncedReadWriterAt(t *testing.T) {
	tests := []struct {
		name               string
		chunkSize          int64
		chunks             int64
		input              []byte
		modifiedInput      []byte
		offset             int64
		expectedData       []byte
		expectedN          int
		expectedReadErr    error
		expectedWriteErr   error
		writeToRemoteFirst bool
		syncThenModify     bool
	}{
		{
			name:               "Write to first chunk and read without error",
			chunkSize:          4,
			chunks:             1,
			input:              []byte("test"),
			offset:             0,
			expectedData:       []byte("test"),
			expectedN:          4,
			expectedReadErr:    nil,
			expectedWriteErr:   nil,
			writeToRemoteFirst: false,
		},
		{
			name:               "Write to second chunk and read without error",
			chunkSize:          4,
			chunks:             2,
			input:              []byte("test"),
			offset:             4,
			expectedData:       []byte("test"),
			expectedN:          4,
			expectedReadErr:    nil,
			expectedWriteErr:   nil,
			writeToRemoteFirst: false,
		},
		{
			name:               "Write to first chunk in remote and read from synced",
			chunkSize:          4,
			chunks:             1,
			input:              []byte("test"),
			offset:             0,
			expectedData:       []byte("test"),
			expectedN:          4,
			expectedReadErr:    nil,
			expectedWriteErr:   nil,
			writeToRemoteFirst: true,
		},
		{
			name:               "Write to second chunk in remote and read from synced",
			chunkSize:          4,
			chunks:             2,
			input:              []byte("test"),
			offset:             4,
			expectedData:       []byte("test"),
			expectedN:          4,
			expectedReadErr:    nil,
			expectedWriteErr:   nil,
			writeToRemoteFirst: true,
		},
		{
			name:               "Sync, modify remote, mark as remote, and resync",
			chunkSize:          4,
			chunks:             2,
			input:              []byte("test"),
			modifiedInput:      []byte("pass"),
			offset:             4,
			expectedData:       []byte("pass"),
			expectedN:          4,
			expectedReadErr:    nil,
			expectedWriteErr:   nil,
			writeToRemoteFirst: true,
			syncThenModify:     true,
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

			localOffsets := map[int64]struct{}{}

			srw := NewSyncedReadWriterAt(
				remote,
				local,
				func(off int64) error {
					localOffsets[off] = struct{}{}

					return nil
				},
			)

			if tc.writeToRemoteFirst {
				wn, werr := remote.WriteAt(tc.input, tc.offset)
				if werr != tc.expectedWriteErr {
					t.Errorf("WriteAt to remote error: expected %v, got %v", tc.expectedWriteErr, werr)
				}

				if wn != tc.expectedN {
					t.Errorf("WriteAt to remote bytes written: expected %v, got %v", tc.expectedN, wn)
				}
			} else {
				wn, werr := srw.WriteAt(tc.input, tc.offset)
				if werr != tc.expectedWriteErr {
					t.Errorf("WriteAt error: expected %v, got %v", tc.expectedWriteErr, werr)
				}

				if wn != tc.expectedN {
					t.Errorf("WriteAt bytes written: expected %v, got %v", tc.expectedN, wn)
				}
			}

			if tc.syncThenModify {
				rbuf := make([]byte, len(tc.input))
				rn, rerr := srw.ReadAt(rbuf, tc.offset)
				if rerr != tc.expectedReadErr {
					t.Errorf("ReadAt error: expected %v, got %v", tc.expectedReadErr, rerr)
				}

				if rn != tc.expectedN {
					t.Errorf("ReadAt bytes read: expected %v, got %v", tc.expectedN, rn)
				}

				_, werr := remote.WriteAt(tc.modifiedInput, tc.offset)
				if werr != nil {
					t.Errorf("WriteAt to remote error: expected %v, got %v", nil, werr)
				}

				srw.MarkAsRemote([]int64{tc.offset})

				rn, rerr = srw.ReadAt(rbuf, tc.offset)
				if rerr != tc.expectedReadErr {
					t.Errorf("ReadAt error after resync: expected %v, got %v", tc.expectedReadErr, rerr)
				}

				if rn != tc.expectedN {
					t.Errorf("ReadAt bytes read after resync: expected %v, got %v", tc.expectedN, rn)
				}

				if !bytes.Equal(rbuf, tc.expectedData) {
					t.Errorf("ReadAt data after resync: expected %v, got %v", tc.expectedData, rbuf)
				}
			} else {
				rbuf := make([]byte, len(tc.input))
				rn, rerr := srw.ReadAt(rbuf, tc.offset)
				if rerr != tc.expectedReadErr {
					t.Errorf("ReadAt error: expected %v, got %v", tc.expectedReadErr, rerr)
				}

				if rn != tc.expectedN {
					t.Errorf("ReadAt bytes read: expected %v, got %v", tc.expectedN, rn)
				}

				if !bytes.Equal(rbuf, tc.expectedData) {
					t.Errorf("ReadAt data: expected %v, got %v", tc.expectedData, rbuf)
				}

				localBuf := make([]byte, len(tc.input))
				_, err = local.ReadAt(localBuf, tc.offset)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(localBuf, tc.input) {
					t.Errorf("Data in local backend did not match expected. got %v, want %v", localBuf, tc.input)
				}

				if _, ok := localOffsets[tc.offset]; !ok {
					t.Errorf("Chunk at offset %d not marked as local", tc.offset)
				}
			}
		})
	}
}
