package chunks

import (
	"context"
	"sync"
	"time"
)

type Pusher struct {
	ctx    context.Context
	cancel context.CancelFunc

	local  ReadWriterAt
	remote ReadWriterAt

	chunkSize int64

	pushableOffsets     map[int64]struct{}
	pushableOffsetsLock sync.Mutex
	changedOffsets      map[int64]*sync.Mutex
	changedOffsetsLock  sync.Mutex

	pushInterval time.Duration
	pushTicker   *time.Ticker

	workerWg  sync.WaitGroup
	workerSem chan struct{}
	errChan   chan error
}

func NewPusher(
	ctx context.Context,
	local, remote ReadWriterAt,
	chunkSize int64,
	pushInterval time.Duration,
) *Pusher {
	ctx, cancel := context.WithCancel(ctx)
	return &Pusher{
		ctx:    ctx,
		cancel: cancel,

		local:  local,
		remote: remote,

		chunkSize: chunkSize,

		pushableOffsets: make(map[int64]struct{}),
		changedOffsets:  make(map[int64]*sync.Mutex),

		pushInterval: pushInterval,

		errChan: make(chan error),
	}
}

func (p *Pusher) Init(workers int64) error {
	p.pushTicker = time.NewTicker(p.pushInterval)
	p.workerSem = make(chan struct{}, workers)
	p.errChan = make(chan error)

	go p.pushChunks()

	return nil
}

func (p *Pusher) MarkOffsetPushable(off int64) error {
	p.pushableOffsetsLock.Lock()
	defer p.pushableOffsetsLock.Unlock()

	p.pushableOffsets[off] = struct{}{}

	return nil
}

func (p *Pusher) pushChunks() {
	for {
		select {
		case <-p.pushTicker.C:
			p.workerWg.Add(1)

			if err := p.Flush(); err != nil {
				p.errChan <- err
			}

			p.workerWg.Done()

		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pusher) Flush() error {
	p.changedOffsetsLock.Lock()

	var wg sync.WaitGroup

	for off, lock := range p.changedOffsets {
		wg.Add(1)

		p.workerSem <- struct{}{}

		go func(off int64, lock *sync.Mutex) {
			defer wg.Done()

			lock.Lock()

			// First fetch from local ReaderAt, then copy to remote one
			b := make([]byte, p.chunkSize)

			if _, err := p.local.ReadAt(b, off); err != nil {
				lock.Unlock()

				p.errChan <- err

				<-p.workerSem

				return
			}

			if _, err := p.remote.WriteAt(b, off); err != nil {
				lock.Unlock()

				p.errChan <- err

				<-p.workerSem

				return
			}

			delete(p.changedOffsets, off)

			lock.Unlock()

			<-p.workerSem
		}(off, lock)
	}

	wg.Wait()

	p.changedOffsetsLock.Unlock()

	return nil
}

func (p *Pusher) Wait() error {
	for err := range p.errChan {
		if err != nil {
			_ = p.Close()

			return err
		}
	}

	return nil
}

func (p *Pusher) Close() error {
	p.Flush()

	p.cancel()
	p.pushTicker.Stop()
	p.workerWg.Wait()

	close(p.errChan)

	return nil
}

func (p *Pusher) ReadAt(b []byte, off int64) (n int, err error) {
	return p.local.ReadAt(b, off)
}

func (p *Pusher) WriteAt(b []byte, off int64) (n int, err error) {
	p.pushableOffsetsLock.Lock()
	_, ok := p.pushableOffsets[off]
	p.pushableOffsetsLock.Unlock()

	if ok {
		p.changedOffsetsLock.Lock()
		if _, exists := p.changedOffsets[off]; !exists {
			p.changedOffsets[off] = &sync.Mutex{}
		}
		p.changedOffsetsLock.Unlock()

		p.changedOffsets[off].Lock()
		defer p.changedOffsets[off].Unlock()
	}

	return p.local.WriteAt(b, off)
}
