package chunks

import (
	"io"
	"sync"
)

type chunk struct {
	lock  sync.Mutex
	local bool
}

type SyncedReadWriterAt struct {
	remote io.ReaderAt
	local  ReadWriterAt

	chunks     map[int64]*chunk
	chunksLock sync.Mutex

	onChunkIsLocal func(off int64) error
}

func NewSyncedReadWriterAt(remote io.ReaderAt, local ReadWriterAt, onChunkIsLocal func(off int64) error) *SyncedReadWriterAt {
	return &SyncedReadWriterAt{
		remote,
		local,

		map[int64]*chunk{},
		sync.Mutex{},

		onChunkIsLocal,
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

	// First fetch from remote ReaderAt, then copy to local one
	n, err = c.remote.ReadAt(p, off)
	if err != nil {
		return 0, err
	}

	if _, err = c.local.WriteAt(p, off); err != nil {
		return 0, err
	}

	chk.local = true

	// Mark the chunk as local so that it gets synced back on the _next_ read
	// We need to mark the chunk _after_ the write to the local file so that
	// we don't mark the writes caused by pulling the chunks as needing to be synced back
	if err := c.onChunkIsLocal(off); err != nil {
		return 0, err
	}

	return n, nil
}

func (c *SyncedReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	chk := c.getOrTrackChunk(off)

	chk.lock.Lock()
	defer chk.lock.Unlock()

	if !chk.local {
		chk.local = true

		// Mark the chunk as local so that it gets synced back on _this_ write
		// Doing so on _this_ write is necessary so that we catch
		// writes to chunks that haven't been pulled yet
		if err := c.onChunkIsLocal(off); err != nil {
			return 0, err
		}
	}

	n, err = c.local.WriteAt(p, off)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c *SyncedReadWriterAt) MarkAsRemote(dirtyOffsets []int64) {
	c.chunksLock.Lock()
	defer c.chunksLock.Unlock()

	for _, off := range dirtyOffsets {
		if chk, ok := c.chunks[off]; ok {
			chk.lock.Lock()
			chk.local = false
			chk.lock.Unlock()
		}
	}
}
