package chunks

import (
	"context"
	"io"
	"sync"
)

type Puller struct {
	backend   io.ReaderAt
	chunkSize int64
	chunks    int64
	workers   int
	errChan   chan error
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewPuller(ctx context.Context, backend io.ReaderAt, chunkSize, chunks int64) *Puller {
	ctx, cancel := context.WithCancel(ctx)

	return &Puller{
		backend:   backend,
		chunkSize: chunkSize,
		chunks:    chunks,
		errChan:   make(chan error),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (p *Puller) Init(workers int) error {
	p.workers = workers
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.pullChunks(i)
	}

	return nil
}

func (p *Puller) pullChunks(workerId int) {
	defer p.wg.Done()

	for i := int64(workerId); i < p.chunks; i += int64(p.workers) {
		select {
		case <-p.ctx.Done():
			return

		default:
			_, err := p.backend.ReadAt(make([]byte, p.chunkSize), i*p.chunkSize)
			if err != nil {
				p.errChan <- err

				return
			}
		}
	}
}

func (p *Puller) Wait() error {
	go func() {
		p.wg.Wait()
		close(p.errChan)
	}()

	for err := range p.errChan {
		if err != nil {
			_ = p.Close()

			return err
		}
	}

	return nil
}

func (p *Puller) Close() error {
	p.cancel()
	p.wg.Wait()

	return nil
}
