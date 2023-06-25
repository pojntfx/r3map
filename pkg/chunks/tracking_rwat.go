package chunks

import (
	"sync"
)

type TrackingReadWriterAt struct {
	backend ReadWriterAt

	tracking          bool
	trackingLock      sync.Mutex
	dirtyOffsetsIndex map[int64]struct{}
	dirtyOffsetsStore []int64 // Separate store from the index so that we don't need to convert it the map to an []int64 later
}

func NewTrackingReadWriterAt(backend ReadWriterAt) *TrackingReadWriterAt {
	return &TrackingReadWriterAt{
		backend:           backend,
		dirtyOffsetsIndex: map[int64]struct{}{},
		dirtyOffsetsStore: []int64{},
	}
}

func (c *TrackingReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	return c.backend.ReadAt(p, off)
}

func (c *TrackingReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	c.trackingLock.Lock()
	defer c.trackingLock.Unlock()

	if c.tracking {
		if _, found := c.dirtyOffsetsIndex[off]; !found {
			c.dirtyOffsetsIndex[off] = struct{}{}
			c.dirtyOffsetsStore = append(c.dirtyOffsetsStore, off)
		}
	}

	return c.backend.WriteAt(p, off)
}

func (c *TrackingReadWriterAt) Sync() []int64 {
	c.trackingLock.Lock()
	defer c.trackingLock.Unlock()

	rv := c.dirtyOffsetsStore

	c.dirtyOffsetsIndex = map[int64]struct{}{}
	c.dirtyOffsetsStore = []int64{}
	c.tracking = false

	return rv
}

func (c *TrackingReadWriterAt) Track() {
	c.trackingLock.Lock()
	defer c.trackingLock.Unlock()

	c.tracking = true
}
