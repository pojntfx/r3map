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
	"syscall"
	"time"
	"unsafe"

	"github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/utils"
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
	workers := flag.Int64("workers", 1, "Puller workers to launch in the background")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

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

	path, err := utils.FindUnusedNBDDevice()
	if err != nil {
		panic(err)
	}

	df, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer df.Close()

	remote := chunks.NewChunkedReadWriterAt(remoteFile, *chunkSize, *chunkCount)
	local := chunks.NewChunkedReadWriterAt(localFile, *chunkSize, *chunkCount)

	srw := chunks.NewSyncedReadWriterAt(remote, local)

	ctx := context.Background()

	puller := chunks.NewPuller(
		ctx,
		srw,
		*chunkSize,
		*chunkCount,
		func(offset int64) int64 {
			return 1
		},
	)

	if err := puller.Init(*workers); err != nil {
		panic(err)
	}
	defer puller.Close()

	go func() {
		if err := puller.Wait(); err != nil {
			panic(err)
		}
	}()

	arw := chunks.NewArbitraryReadWriterAt(srw, *chunkSize)

	b := backend.NewReaderAtBackend(
		arw,
		func() (int64, error) {
			stat, err := remoteFile.Stat()
			if err != nil {
				return 0, err
			}

			return stat.Size(), nil
		},
		localFile.Sync,
		*verbose,
	)

	d := device.NewDevice(
		b,
		df,

		nil,
		nil,
	)

	if err := d.Open(); err != nil {
		panic(err)
	}
	defer d.Close()

	go func() {
		if err := d.Wait(); err != nil {
			panic(err)
		}
	}()

	cf, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer cf.Close()

	size, err := cf.Seek(0, io.SeekEnd)
	if err != nil {
		panic(err)
	}

	p, err := syscall.Mmap(
		int(cf.Fd()),
		0,
		int(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		panic(err)
	}
	defer syscall.Munmap(p)

	defer func() {
		_, _, _ = syscall.Syscall(
			syscall.SYS_MSYNC,
			uintptr(unsafe.Pointer(&p[0])),
			uintptr(len(p)),
			uintptr(syscall.MS_SYNC),
		)
	}()

	before := time.Now()

	localHash, err := getHash(bytes.NewReader(p))
	if err != nil {
		panic(err)
	}

	after := time.Since(before)

	if _, err := remoteFile.Seek(0, io.SeekStart); err != nil {
		panic(err)
	}

	remoteHash, err := getHash(remoteFile)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(remoteHash, localHash) {
		panic("Remote and local hashes don't match")
	}

	fmt.Printf("%.2f MB/s\n", float64(*chunkSize**chunkCount)/(1024*1024)/after.Seconds())
}
