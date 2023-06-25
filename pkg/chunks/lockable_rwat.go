package chunks

import "sync"

type LockableReadWriterAt struct {
	backend ReadWriterAt
	lock    *sync.Cond
	locked  bool
}

func NewLockableReadWriterAt(backend ReadWriterAt) *LockableReadWriterAt {
	return &LockableReadWriterAt{
		backend: backend,
		lock:    sync.NewCond(&sync.Mutex{}),
		locked:  true,
	}
}

func (a *LockableReadWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	a.lock.L.Lock()
	if a.locked {
		a.lock.Wait()
	}
	a.lock.L.Unlock()

	return a.backend.ReadAt(p, off)
}

func (a *LockableReadWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	a.lock.L.Lock()
	if a.locked {
		a.lock.Wait()
	}
	a.lock.L.Unlock()

	return a.backend.WriteAt(p, off)
}

func (a *LockableReadWriterAt) Unlock() {
	a.lock.L.Lock()
	a.locked = false
	a.lock.Broadcast()
	a.lock.L.Unlock()
}
