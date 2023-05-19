package backend

import (
	"context"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisBackend struct {
	ctx context.Context

	redisClient *redis.Client
	size        int64

	verbose bool
}

func NewRedisBackend(
	ctx context.Context,
	redisOptions *redis.Options,
	size int64,
	verbose bool,
) *RedisBackend {
	return &RedisBackend{
		ctx: ctx,

		redisClient: redis.NewClient(redisOptions),
		size:        size,

		verbose: verbose,
	}
}

func (b *RedisBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	val, err := b.redisClient.Get(b.ctx, strconv.FormatInt(off, 10)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return len(p), nil
		}

		return 0, err
	}

	return copy(p, val), nil
}

func (b *RedisBackend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	if err := b.redisClient.Set(b.ctx, strconv.FormatInt(off, 10), p, 0).Err(); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (b *RedisBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.size, nil
}

func (b *RedisBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	return nil
}
