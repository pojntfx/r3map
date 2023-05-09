package chunks

import (
	"io"
)

type ArbitraryReadWriterAt struct {
	crw       *ChunkedReadWriterAt
	chunkSize int64
}

func NewArbitraryReadWriterAt(crw *ChunkedReadWriterAt, chunkSize int64) *ArbitraryReadWriterAt {
	return &ArbitraryReadWriterAt{
		crw:       crw,
		chunkSize: chunkSize,
	}
}

func (a *ArbitraryReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	totalRead := 0
	remaining := len(p)

	for remaining > 0 {
		chunkIndex := off / a.chunkSize
		indexedOffset := off % a.chunkSize
		readSize := int64(min(remaining, int(a.chunkSize-indexedOffset)))

		buf := make([]byte, a.chunkSize)
		_, err := a.crw.ReadAt(buf, chunkIndex*a.chunkSize)
		if err != nil && err != io.EOF {
			return totalRead, err
		}

		copy(p[totalRead:], buf[indexedOffset:indexedOffset+readSize])

		totalRead += int(readSize)
		remaining -= int(readSize)
		off += readSize
	}

	return totalRead, nil
}

func (a *ArbitraryReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	totalWritten := 0
	remaining := len(p)

	for remaining > 0 {
		chunkIndex := off / a.chunkSize
		indexedOffset := off % a.chunkSize
		writeSize := int(min(remaining, int(a.chunkSize-indexedOffset)))

		buf := make([]byte, a.chunkSize)
		_, err := a.crw.ReadAt(buf, chunkIndex*a.chunkSize) // Read the existing chunk
		if err != nil && (err != io.EOF || indexedOffset != 0) {
			return totalWritten, err
		}

		// Modify the chunk with the provided data
		copy(buf[indexedOffset:], p[totalWritten:totalWritten+writeSize])

		// Write back the updated chunk
		_, err = a.crw.WriteAt(buf, chunkIndex*a.chunkSize)
		if err != nil {
			return totalWritten, err
		}

		totalWritten += writeSize
		remaining -= writeSize
		off += int64(writeSize)
	}

	return totalWritten, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
