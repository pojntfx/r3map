package chunks

import (
	"os"
	"testing"
	"time"
)

func TestLockableReadWriterAt(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		read   bool
		unlock bool
	}{
		{
			name:   "Can write when unlocked",
			input:  []byte("1234"),
			read:   false,
			unlock: true,
		},
		{
			name:   "Writes time out when locked",
			input:  []byte("1234"),
			read:   false,
			unlock: false,
		},
		{
			name:   "Can read when unlocked",
			input:  []byte("1234"),
			read:   true,
			unlock: true,
		},
		{
			name:   "Reads time out when locked",
			input:  []byte("1234"),
			read:   true,
			unlock: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(f.Name())

			lrw := NewLockableReadWriterAt(f)

			if tc.unlock {
				lrw.Unlock()
			}

			if tc.input != nil {
				done := make(chan struct{})

				go func() {
					wn, werr := lrw.WriteAt(tc.input, 0)
					close(done)

					if werr != nil {
						t.Error(err)
					}

					if wn != len(tc.input) {
						t.Errorf("WriteAt bytes written: expected %v, got %v", len(tc.input), wn)
					}
				}()

				timeout := time.After(1 * time.Millisecond)

				select {
				case <-done:
					if !tc.unlock {
						t.Error("WriteAt should not complete when locked")
					}
				case <-timeout:
					if tc.unlock {
						t.Error("WriteAt timed out when unlocked")
					}
				}
			}

			if tc.read {
				done := make(chan struct{})

				go func() {
					wn, werr := lrw.ReadAt(tc.input, 0)
					close(done)

					if werr != nil {
						t.Error(err)
					}

					if wn != len(tc.input) {
						t.Errorf("ReadAt bytes read: expected %v, got %v", len(tc.input), wn)
					}
				}()

				timeout := time.After(1 * time.Millisecond)

				select {
				case <-done:
					if !tc.unlock {
						t.Error("ReadAt should not complete when locked")
					}
				case <-timeout:
					if tc.unlock {
						t.Error("ReadAt timed out when unlocked")
					}
				}
			}
		})
	}
}

func TestLockableReadWriterAtWithGenericTest(t *testing.T) {
	TestArbitraryReadWriterAtGeneric(
		t,
		func(chunkSize, chunkCount int64) (ReadWriterAt, func() error, error) {
			f, err := os.CreateTemp("", "")
			if err != nil {
				return nil, nil, err
			}

			if err := f.Truncate(chunkSize * chunkCount); err != nil {
				return nil, nil, err
			}

			l := NewLockableReadWriterAt(f)

			l.Unlock()

			return l,
				func() error {
					return os.RemoveAll(f.Name())
				},
				nil
		},
		[]ReadWriteConfiguration{
			{
				ChunkSizes:  []int64{2, 4, 8},
				ChunkCount:  16,
				BufferSizes: []int64{1, 2, 4, 8, 16},
			},
			{
				ChunkSizes:  []int64{2, 4, 8},
				ChunkCount:  64,
				BufferSizes: []int64{1, 2, 4, 8, 16, 32, 64, 128},
			},
		},
	)
}
