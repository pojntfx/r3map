package chunks

import (
	"context"
	"io"
	"sort"
	"sync"
)

type Puller struct {
	backend io.ReaderAt

	chunkSize    int64
	chunks       int64
	finalized    bool
	finalizeCond *sync.Cond

	errs   chan error
	ctx    context.Context
	cancel context.CancelFunc

	workersWg sync.WaitGroup

	chunkIndexes              []int64
	chunkIndexesLock          sync.Mutex
	nextChunk                 int64
	nextChunkAndFinalizedLock sync.Mutex

	closeLock sync.Mutex
}

func NewPuller(
	ctx context.Context,
	backend io.ReaderAt,
	chunkSize, chunks int64,
	pullPriority func(off int64) int64,
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

	puller.finalizeCond = sync.NewCond(&sync.Mutex{})

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
	p.nextChunkAndFinalizedLock.Lock()

	if !p.finalized && p.nextChunk >= p.chunks-1 {
		p.finalizeCond.L.Lock()

		p.nextChunkAndFinalizedLock.Unlock()
		p.finalizeCond.Wait()
		p.nextChunkAndFinalizedLock.Lock()

		p.finalizeCond.L.Unlock()
	}

	nextChunk := p.nextChunk
	p.nextChunk++

	p.nextChunkAndFinalizedLock.Unlock()

	return nextChunk
}

func (p *Puller) pullChunks() {
	defer p.workersWg.Done()

	for {
		chunk := p.getNextChunk()

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

func (p *Puller) Finalize(dirtyOffsets []int64) {
	p.nextChunkAndFinalizedLock.Lock()
	defer p.nextChunkAndFinalizedLock.Unlock()

	if p.finalized {
		return
	}

	p.chunkIndexesLock.Lock()
	defer p.chunkIndexesLock.Unlock()

	// Calculate dirty indexes once and store them
	dirtyIndexes := make([]int64, len(dirtyOffsets))
	for i, dirtyOffset := range dirtyOffsets {
		dirtyIndexes[i] = dirtyOffset / p.chunkSize
	}

	// Create new slice with all elements before the insertion point
	newChunkIndexes := make([]int64, p.nextChunk+1)
	copy(newChunkIndexes, p.chunkIndexes[:p.nextChunk+1])

	// Append dirty indexes
	newChunkIndexes = append(newChunkIndexes, dirtyIndexes...)

	// Append the remaining elements
	newChunkIndexes = append(newChunkIndexes, p.chunkIndexes[p.nextChunk+1:]...)

	p.chunkIndexes = newChunkIndexes

	p.chunks += int64(len(dirtyOffsets))

	p.finalized = true

	p.finalizeCond.L.Lock()
	p.finalizeCond.Broadcast()
	p.finalizeCond.L.Unlock()
}

func (p *Puller) Wait() error {
	go func() {
		p.workersWg.Wait()

		p.closeLock.Lock()
		defer p.closeLock.Unlock()

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
