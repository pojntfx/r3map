package chunks

import (
	"context"
	"io"
	"sync"
	"time"
)

type Pusher struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	local               io.ReadWriterAt
	remote              io.ReadWriterAt
	chunkSize           int64
	pushInterval        time.Duration
	pushableOffsets     map[int64]struct{}
	pushableOffsetsLock sync.Mutex
	changedOffsets      map[int64]struct{}
	changedOffsetsLock  sync.Mutex
	ticker              *time.Ticker
	workerWg            sync.WaitGroup
	errChan             chan error
}

func NewPusher(
	ctx context.Context,
	local, remote io.ReadWriterAt,
	chunkSize int64,
	pushInterval time.Duration,
) *Pusher {
	ctx, cancel := context.WithCancel(ctx)
	return &Pusher{
		ctx:             ctx,
		cancel:          cancel,
		local:           local,
		remote:          remote,
		chunkSize:       chunkSize,
		pushInterval:    pushInterval,
		pushableOffsets: make(map[int64]struct{}),
		changedOffsets:  make(map[int64]struct{}),
		errChan:         make(chan error),
	}
}

func (p *Pusher) Init() error {
	p.ticker = time.NewTicker(p.pushInterval)
	go p.pushChunks()
	return nil
}

func (p *Pusher) pushChunks() {
	for {
		select {
		case <-p.ticker.C:
			p.workerWg.Add(1)
			err := p.Flush()
			if err != nil {
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
		// read from local
		buf := make([]byte, p.chunkSize)
		_, err := p.local.ReadAt(buf, off)
		if err != nil {
			return err
		}

		// write to remote
		_, err = p.remote.WriteAt(buf, off)
		if err != nil {
			return err
		}

		// remove offset from changedOffsets
		delete(p.changedOffsets, off)
	}

	return nil
}

func (p *Pusher) MarkOffsetPushable(off int64) error {
	p.pushableOffsetsLock.Lock()
	defer p.pushableOffsetsLock.Unlock()

	p.pushableOffsets[off] = struct{}{}
	return nil
}

func (p *Pusher) ReadAt(b []byte, off int64) (n int, err error) {
	return p.local.ReadAt(b, off)
}
