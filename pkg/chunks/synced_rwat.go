package chunks

import (
	"sync"
)

type chunk struct {
	lock  sync.Mutex
	local bool
}

type SyncedReadWriterAt struct {
	remote ReadWriterAt
	local  ReadWriterAt

	chunks     map[int64]*chunk
	chunksLock sync.Mutex
}

func NewSyncedReadWriterAt(remote ReadWriterAt, local ReadWriterAt) *SyncedReadWriterAt {
	return &SyncedReadWriterAt{
		remote,
		local,

		map[int64]*chunk{},
		sync.Mutex{},
	}
}

func (c *SyncedReadWriterAt) getOrTrackChunk(off int64) *chunk {
	c.chunksLock.Lock()
	chk, ok := c.chunks[off]
	if !ok {
		chk = &chunk{
			local: false,
			lock:  sync.Mutex{},
		}
		c.chunks[off] = chk
	}
	c.chunksLock.Unlock()

	return chk
}

func (c *SyncedReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	chk := c.getOrTrackChunk(off)

	chk.lock.Lock()
	defer chk.lock.Unlock()

	if chk.local {
		// Return from local ReaderAt
		n, err = c.local.ReadAt(p, off)
		if err != nil {
			return 0, err
		}

		return n, nil
	}

	// First fetch from local ReaderAt, then copy to local one
	n, err = c.remote.ReadAt(p, off)
	if err != nil {
		return 0, err
	}

	n, err = c.local.WriteAt(p, off)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c *SyncedReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	chk := c.getOrTrackChunk(off)

	chk.lock.Lock()
	defer chk.lock.Unlock()

	chk.local = true

	n, err = c.local.WriteAt(p, off)
	if err != nil {
		return 0, err
	}

	return n, nil
}
