package backend

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type DirectoryBackend struct {
	directory string
	size      int64
	chunkSize int64

	files     map[int64]*os.File
	filesLock sync.Mutex

	verbose bool
}

func NewDirectoryBackend(
	directory string,
	size int64,
	chunkSize int64,
	verbose bool,
) *DirectoryBackend {
	return &DirectoryBackend{
		directory: directory,
		size:      size,
		chunkSize: chunkSize,

		files: map[int64]*os.File{},

		verbose: verbose,
	}
}

func (b *DirectoryBackend) getOrTrackFile(off int64) (*os.File, error) {
	b.filesLock.Lock()
	f, ok := b.files[off]
	if !ok {
		var err error
		f, err = os.OpenFile(filepath.Join(b.directory, strconv.FormatInt(off, 10)), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			b.filesLock.Unlock()

			return nil, err
		}

		if err := f.Truncate(b.chunkSize); err != nil {
			b.filesLock.Unlock()

			return nil, err
		}

		b.files[off] = f
	}
	b.filesLock.Unlock()

	return f, nil
}

func (b *DirectoryBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	f, err := b.getOrTrackFile(off)
	if err != nil {
		return -1, err
	}

	return f.ReadAt(p, 0)
}

func (b *DirectoryBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	f, err := b.getOrTrackFile(off)
	if err != nil {
		return -1, err
	}

	return f.WriteAt(p, 0)
}

func (b *DirectoryBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.size, nil
}

func (b *DirectoryBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	b.filesLock.Lock()
	defer b.filesLock.Unlock()

	for _, f := range b.files {
		if err := f.Sync(); err != nil {
			return err
		}
	}

	return nil
}
