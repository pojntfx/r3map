package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pojntfx/r3map/pkg/chunks"
)

func getHash(file io.Reader) ([]byte, error) {
	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return []byte{}, err
	}

	return hash.Sum(nil), nil
}

func main() {
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	chunkCount := flag.Int64("chunk-count", 8192, "Amount of chunks to create")
	workers := flag.Int64("workers", 1, "Workers to launch in the background")

	flag.Parse()

	remoteFile, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(remoteFile.Name())

	if _, err := io.CopyN(remoteFile, rand.Reader, *chunkCount**chunkSize); err != nil {
		panic(err)
	}

	localFile, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(localFile.Name())

	if err := localFile.Truncate(*chunkSize * *chunkCount); err != nil {
		panic(err)
	}

	remote := chunks.NewChunkedReadWriterAt(remoteFile, *chunkSize, *chunkCount)
	local := chunks.NewChunkedReadWriterAt(localFile, *chunkSize, *chunkCount)

	srw := chunks.NewSyncedReadWriterAt(remote, local)

	ctx := context.Background()

	puller := chunks.NewPuller(ctx, srw, *chunkSize, *chunkCount)

	before := time.Now()
	if err := puller.Init(*workers); err != nil {
		panic(err)
	}
	defer puller.Close()

	if err := puller.Wait(); err != nil {
		panic(err)
	}
	after := time.Since(before)

	if err := puller.Close(); err != nil {
		panic(err)
	}

	if _, err := remoteFile.Seek(0, io.SeekStart); err != nil {
		panic(err)
	}

	if _, err := localFile.Seek(0, io.SeekStart); err != nil {
		panic(err)
	}

	remoteHash, err := getHash(remoteFile)
	if err != nil {
		panic(err)
	}

	localHash, err := getHash(localFile)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(remoteHash, localHash) {
		panic("Remote and local hashes don't match")
	}

	fmt.Printf("%.2f MB/s\n", float64(*chunkSize**chunkCount)/(1024*1024)/after.Seconds())
}
