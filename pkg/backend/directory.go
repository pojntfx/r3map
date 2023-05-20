package backend

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type fileWithLock struct {
	file *os.File
	lock *sync.Mutex
}

type DirectoryBackend struct {
	directory    string
	size         int64
	chunkSize    int64
	maxOpenFiles int64

	files      map[int64]*fileWithLock
	filesLock  sync.Mutex
	filesQueue []int64

	verbose bool
}

func NewDirectoryBackend(
	directory string,
	size int64,
	chunkSize int64,
	maxOpenFiles int64,
	verbose bool,
) *DirectoryBackend {
	return &DirectoryBackend{
		directory:    directory,
		size:         size,
		chunkSize:    chunkSize,
		maxOpenFiles: maxOpenFiles,

		files:      map[int64]*fileWithLock{},
		filesQueue: make([]int64, 0, maxOpenFiles),

		verbose: verbose,
	}
}

func (b *DirectoryBackend) getOrTrackFile(off int64, writing bool) (*fileWithLock, error) {
	b.filesLock.Lock()
	defer b.filesLock.Unlock()

	fl, ok := b.files[off]
	if !ok {
		if int64(len(b.filesQueue)) == b.maxOpenFiles {
			lruOff := b.filesQueue[0]
			lruFl := b.files[lruOff]

			lruFl.lock.Lock()

			if lruFl.file != nil {
				if err := lruFl.file.Close(); err != nil {
					lruFl.lock.Unlock()

					return nil, err
				}

				lruFl.file = nil
			}

			lruFl.lock.Unlock()

			delete(b.files, lruOff)
			b.filesQueue = b.filesQueue[1:]
		}

		f, err := os.OpenFile(filepath.Join(b.directory, strconv.FormatInt(off, 10)), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return nil, err
		}

		if !writing {
			if err := f.Truncate(b.chunkSize); err != nil {
				return nil, err
			}
		}

		fl = &fileWithLock{
			file: f,
			lock: &sync.Mutex{},
		}

		b.files[off] = fl
		b.filesQueue = append(b.filesQueue, off)
	}

	return fl, nil
}

func (b *DirectoryBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	f, err := b.getOrTrackFile(off, false)
	if err != nil {
		return -1, err
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	return f.file.ReadAt(p, 0)
}

func (b *DirectoryBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	f, err := b.getOrTrackFile(off, true)
	if err != nil {
		return -1, err
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	return f.file.WriteAt(p, 0)
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
		if err := f.file.Sync(); err != nil {
			return err
		}
	}

	return nil
}
