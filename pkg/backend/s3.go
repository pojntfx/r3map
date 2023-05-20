package backend

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"strconv"

	"github.com/minio/minio-go"
)

var (
	errNoSuchKey = errors.New("The specified key does not exist.") // Minio doesn't export errors
)

type S3Backend struct {
	ctx context.Context

	client *minio.Client
	bucket string
	prefix string
	size   int64

	verbose bool
}

func NewS3Backend(
	ctx context.Context,
	client *minio.Client,
	bucket string,
	prefix string,
	size int64,
	verbose bool,
) *S3Backend {
	return &S3Backend{
		ctx: ctx,

		client: client,
		bucket: bucket,
		prefix: prefix,
		size:   size,

		verbose: verbose,
	}
}

func (b *S3Backend) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	obj, err := b.client.GetObject(b.bucket, b.prefix+"-"+strconv.FormatInt(off, 10), minio.GetObjectOptions{})
	if err != nil {
		if err.Error() == errNoSuchKey.Error() {
			return len(p), nil
		}

		return 0, err
	}

	nn, err := obj.Read(p)
	if err != nil && nn == len(p) && err != io.EOF {
		if err.Error() == errNoSuchKey.Error() {
			return len(p), nil
		}

		return 0, err
	}

	return int(nn), nil
}

func (b *S3Backend) WriteAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("WriteAt(len(p) = %v, off = %v)", len(p), off)
	}

	nn, err := b.client.PutObject(b.bucket, b.prefix+"-"+strconv.FormatInt(off, 10), bytes.NewReader(p), int64(len(p)), minio.PutObjectOptions{})

	return int(nn), err
}

func (b *S3Backend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.size, nil
}

func (b *S3Backend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	return nil
}
