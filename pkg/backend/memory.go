package backend

import (
	"io"
	"log"
	"sync"
)

type MemoryBackend struct {
	memory  []byte
	lock    sync.Mutex
	verbose bool
}

func NewMemoryBackend(memory []byte, verbose bool) *MemoryBackend {
	return &MemoryBackend{memory, sync.Mutex{}, verbose}
}

func (b *MemoryBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	b.lock.Lock()

	if off >= int64(len(b.memory)) {
		return 0, io.EOF
	}

	n = copy(p, b.memory[off:off+int64(len(p))])

	b.lock.Unlock()

	return
}

func (b *MemoryBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	b.lock.Lock()

	if off >= int64(len(b.memory)) {
		return 0, io.EOF
	}

	n = copy(b.memory[off:off+int64(len(p))], p)

	if n < len(p) {
		return n, io.ErrShortWrite
	}

	b.lock.Unlock()

	return
}

func (b *MemoryBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return int64(len(b.memory)), nil
}

func (b *MemoryBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	return nil
}
