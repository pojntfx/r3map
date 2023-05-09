package chunks

import "io"

type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

type ArbitraryReadWriterAt struct {
	backend   ReadWriterAt
	chunkSize int64
}

func NewArbitraryReadWriterAt(
	backend ReadWriterAt,
	chunkSize int64,
) *ArbitraryReadWriterAt {
	return &ArbitraryReadWriterAt{
		backend,
		chunkSize,
	}
}

func (a *ArbitraryReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	chunkOffset := off % a.chunkSize
	alignedOffset := off - chunkOffset
	buf := make([]byte, a.chunkSize)

	for remaining := int64(len(p)); remaining > 0; {
		remainingChunk := a.chunkSize - chunkOffset

		todo := remaining
		if todo > remainingChunk {
			todo = remainingChunk
		}

		if chunkOffset > 0 || todo < a.chunkSize {
			_, err := a.backend.ReadAt(buf, alignedOffset)
			if err != nil && err != io.EOF {
				return n, err
			}
		}

		copy(buf[chunkOffset:chunkOffset+todo], p[n:n+int(todo)])

		_, err := a.backend.WriteAt(buf, alignedOffset)
		if err != nil {
			return n, err
		}

		n += int(todo)
		remaining -= todo
		chunkOffset = 0
		alignedOffset += a.chunkSize
	}

	return n, nil
}

func (a *ArbitraryReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	chunkOffset := off % a.chunkSize
	alignedOffset := off - chunkOffset
	buf := make([]byte, a.chunkSize)

	for remaining := int64(len(p)); remaining > 0; {
		remainingChunk := a.chunkSize - chunkOffset

		todo := remaining
		if todo > remainingChunk {
			todo = remainingChunk
		}

		_, err := a.backend.ReadAt(buf, alignedOffset)
		if err != nil {
			return n, err
		}

		copy(p[n:n+int(todo)], buf[chunkOffset:chunkOffset+todo])

		n += int(todo)
		remaining -= todo
		chunkOffset = 0
		alignedOffset += a.chunkSize
	}

	return n, nil
}
