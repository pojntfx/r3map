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
	changedOffsets      map[int64]struct{}
	changedOffsetsLock  sync.Mutex

	pushInterval time.Duration
	pushTicker   *time.Ticker

	workerWg sync.WaitGroup
	errChan  chan error
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
		changedOffsets:  make(map[int64]struct{}),

		pushInterval: pushInterval,

		errChan: make(chan error),
	}
}

func (p *Pusher) Init() error {
	p.pushTicker = time.NewTicker(p.pushInterval)

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
	defer p.changedOffsetsLock.Unlock()

	for off := range p.changedOffsets {
		// First fetch from local ReaderAt, then copy to remote one
		b := make([]byte, p.chunkSize)

		if _, err := p.local.ReadAt(b, off); err != nil {
			return err
		}

		if _, err := p.remote.WriteAt(b, off); err != nil {
			return err
		}

		delete(p.changedOffsets, off)
	}

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
		p.changedOffsets[off] = struct{}{}
		p.changedOffsetsLock.Unlock()
	}

	return p.local.WriteAt(b, off)
}
