package chunks

import "sync"

type TrackingReadWriterAt struct {
	backend ReadWriterAt

	tracking     bool
	trackingLock sync.Mutex

	dirtyOffsets map[int64]struct{}
}

func NewTrackingReadWriterAt(backend ReadWriterAt) *TrackingReadWriterAt {
	return &TrackingReadWriterAt{
		backend:      backend,
		dirtyOffsets: map[int64]struct{}{},
	}
}

func (c *TrackingReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	return c.backend.ReadAt(p, off)
}

func (c *TrackingReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	c.trackingLock.Lock()
	defer c.trackingLock.Unlock()

	if c.tracking {
		c.dirtyOffsets[off] = struct{}{}
	}

	return c.backend.WriteAt(p, off)
}

func (c *TrackingReadWriterAt) Sync() []int64 {
	c.trackingLock.Lock()
	defer c.trackingLock.Unlock()

	rv := make([]int64, 0, len(c.dirtyOffsets))
	for off := range c.dirtyOffsets {
		rv = append(rv, off)
	}

	c.dirtyOffsets = map[int64]struct{}{}
	c.tracking = false

	return rv
}

func (c *TrackingReadWriterAt) Track() {
	c.trackingLock.Lock()
	defer c.trackingLock.Unlock()

	c.tracking = true
}
