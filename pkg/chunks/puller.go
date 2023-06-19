package chunks

import (
	"context"
	"io"
	"sort"
	"sync"
)

type Puller struct {
	backend io.ReaderAt

	chunkSize        int64
	chunks           int64
	finalized        bool
	finalizePullLock sync.Mutex
	finalizePullCond *sync.Cond

	errs   chan error
	ctx    context.Context
	cancel context.CancelFunc

	workersWg sync.WaitGroup

	chunkIndexes     []int64
	chunkIndexesLock sync.Mutex
	nextChunk        int64
	nextChunkLock    sync.Mutex
}

func NewPuller(
	ctx context.Context,
	backend io.ReaderAt,
	chunkSize, chunks int64,
	pullPriority func(offset int64) int64,
) *Puller {
	ctx, cancel := context.WithCancel(ctx)

	chunkIndexes := make([]int64, 0, chunks)
	for i := int64(0); i < chunks; i++ {
		chunkIndexes = append(chunkIndexes, i)
	}

	// Sort the chunks according to a pull priority heuristic
	sort.Slice(chunkIndexes, func(a, b int) bool {
		return pullPriority(chunkIndexes[a]) > pullPriority(chunkIndexes[b])
	})

	puller := &Puller{
		backend: backend,

		chunkSize: chunkSize,
		chunks:    chunks,

		errs:   make(chan error),
		ctx:    ctx,
		cancel: cancel,

		chunkIndexes: chunkIndexes,
	}

	puller.finalizePullCond = sync.NewCond(&puller.finalizePullLock)

	return puller
}

func (p *Puller) Open(workers int64) error {
	for i := int64(0); i < workers; i++ {
		p.workersWg.Add(1)

		go p.pullChunks()
	}

	return nil
}

func (p *Puller) getNextChunk() int64 {
	p.nextChunkLock.Lock()
	defer p.nextChunkLock.Unlock()

	chunk := p.nextChunk
	p.nextChunk++

	return chunk
}

func (p *Puller) pullChunks() {
	defer p.workersWg.Done()

	for {
		chunk := p.getNextChunk()

		p.finalizePullLock.Lock()
		for !p.finalized && chunk >= p.chunks {
			p.finalizePullCond.Wait()
			chunk = p.getNextChunk()
		}
		p.finalizePullLock.Unlock()

		if chunk >= p.chunks {
			break
		}

		p.chunkIndexesLock.Lock()
		chunkIndex := p.chunkIndexes[chunk]
		p.chunkIndexesLock.Unlock()

		select {
		case <-p.ctx.Done():
			return

		default:
			_, err := p.backend.ReadAt(make([]byte, p.chunkSize), chunkIndex*p.chunkSize)
			if err != nil {
				p.errs <- err

				return
			}
		}
	}
}

func (p *Puller) FinalizePull(dirtyOffsets []int64) {
	p.finalizePullLock.Lock()
	defer p.finalizePullLock.Unlock()

	p.nextChunkLock.Lock()
	defer p.nextChunkLock.Unlock()

	p.chunkIndexesLock.Lock()
	defer p.chunkIndexesLock.Unlock()

	for _, dirtyOffset := range dirtyOffsets {
		dirtyIndex := dirtyOffset / p.chunkSize
		p.chunkIndexes = append(p.chunkIndexes[:p.nextChunk], append([]int64{dirtyIndex}, p.chunkIndexes[p.nextChunk:]...)...)
	}

	p.chunks += int64(len(dirtyOffsets))
	p.finalized = true
	p.finalizePullCond.Broadcast()
}

func (p *Puller) Wait() error {
	go func() {
		p.workersWg.Wait()

		if p.errs != nil {
			close(p.errs)

			p.errs = nil
		}
	}()

	for err := range p.errs {
		if err != nil {
			_ = p.Close()

			return err
		}
	}

	return nil
}

func (p *Puller) Close() error {
	p.cancel()
	p.workersWg.Wait()

	return nil
}
