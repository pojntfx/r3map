package chunks

import (
	"context"
	"io"
	"sort"
	"sync"
)

type Puller struct {
	backend io.ReaderAt

	chunkSize int64
	chunks    int64

	errChan chan error
	ctx     context.Context
	cancel  context.CancelFunc

	workersWg sync.WaitGroup

	chunkIndexes []int64

	nextChunk     int64
	nextChunkLock sync.Mutex
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

	return &Puller{
		backend: backend,

		chunkSize: chunkSize,
		chunks:    chunks,

		errChan: make(chan error),
		ctx:     ctx,
		cancel:  cancel,

		chunkIndexes: chunkIndexes,
	}
}

func (p *Puller) Init(workers int64) error {
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
		if chunk >= p.chunks {
			break
		}

		chunkIndex := p.chunkIndexes[chunk]

		select {
		case <-p.ctx.Done():
			return

		default:
			_, err := p.backend.ReadAt(make([]byte, p.chunkSize), chunkIndex*p.chunkSize)
			if err != nil {
				p.errChan <- err

				return
			}
		}
	}
}

func (p *Puller) Wait() error {
	go func() {
		p.workersWg.Wait()
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
	p.workersWg.Wait()

	return nil
}
