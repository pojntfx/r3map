package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/utils"
)

type sliceRwat struct {
	data []byte
	rtt  time.Duration
}

func (rw *sliceRwat) ReadAt(p []byte, off int64) (n int, err error) {
	if rw.rtt > 0 {
		time.Sleep(rw.rtt)
	}

	if off >= int64(len(rw.data)) {
		return 0, io.EOF
	}

	n = copy(p, rw.data[off:])

	return n, nil
}

func (rw *sliceRwat) WriteAt(p []byte, off int64) (n int, err error) {
	if rw.rtt > 0 {
		time.Sleep(rw.rtt)
	}

	if off >= int64(len(rw.data)) {
		return 0, io.ErrShortWrite
	}

	n = copy(rw.data[off:], p)
	if n < len(p) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

func allocateSlice(size int) ([]byte, func() error, error) {
	p, err := syscall.Mmap(
		-1,
		0,
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
	if err != nil {
		return []byte{}, nil, nil
	}

	return p, func() error {
		_, _, _ = syscall.Syscall(
			syscall.SYS_MSYNC,
			uintptr(unsafe.Pointer(&p[0])),
			uintptr(len(p)),
			uintptr(syscall.MS_SYNC),
		)

		if err := syscall.Munmap(p); err != nil {
			panic(err)
		}

		return nil
	}, nil
}

func main() {
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	chunkCount := flag.Int64("chunk-count", 8192, "Amount of chunks to create")
	workers := flag.Int64("workers", 512, "Puller workers to launch in the background; pass in 0 to disable preemptive pull")
	completePull := flag.Bool("complete-pull", false, "Whether to completely pull the remote to the local slice before starting benchmark")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	check := flag.Bool("check", true, "Check if local and remote hashes match")
	remoteRTT := flag.Duration("remote-rtt", 0, "Simulated RTT of the remote slice")
	localRTT := flag.Duration("local-rtt", 0, "Simulated RTT of the local slice")

	flag.Parse()

	// Create remote and local slices
	remoteSlice, freeRemoteSlice, err := allocateSlice(int(*chunkSize * *chunkCount))
	if err != nil {
		panic(err)
	}
	defer freeRemoteSlice()

	if _, err := rand.Read(remoteSlice); err != nil {
		panic(err)
	}

	remoteFile := &sliceRwat{remoteSlice, *remoteRTT}

	localSlice, freeLocalSlice, err := allocateSlice(int(*chunkSize * *chunkCount))
	if err != nil {
		panic(err)
	}
	defer freeLocalSlice()

	localFile := &sliceRwat{localSlice, *localRTT}

	// Setup the device
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

	// Setup the puller
	if *workers > 0 {
		puller := chunks.NewPuller(
			ctx,
			srw,
			*chunkSize,
			*chunkCount,
			func(offset int64) int64 {
				return 1
			},
		)

		if !*completePull {
			go func() {
				if err := puller.Wait(); err != nil {
					panic(err)
				}
			}()
		}

		if err := puller.Init(*workers); err != nil {
			panic(err)
		}
		defer puller.Close()

		if *completePull {
			if err := puller.Wait(); err != nil {
				panic(err)
			}
		}
	}

	arw := chunks.NewArbitraryReadWriterAt(srw, *chunkSize)

	b := backend.NewReaderAtBackend(
		arw,
		func() (int64, error) {
			return *chunkSize * *chunkCount, nil
		},
		func() error {
			return nil
		},
		*verbose,
	)

	d := device.NewDevice(
		b,
		df,

		nil,
		nil,
	)

	go func() {
		if err := d.Wait(); err != nil {
			panic(err)
		}
	}()

	if err := d.Open(); err != nil {
		panic(err)
	}
	defer d.Close()

	cf, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer cf.Close()

	// Setup the r3mapped and output slices
	r3mappedSlice, err := syscall.Mmap(
		int(cf.Fd()),
		0,
		int(*chunkSize**chunkCount),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		panic(err)
	}
	defer syscall.Munmap(r3mappedSlice)

	defer func() {
		_, _, _ = syscall.Syscall(
			syscall.SYS_MSYNC,
			uintptr(unsafe.Pointer(&r3mappedSlice[0])),
			uintptr(len(r3mappedSlice)),
			uintptr(syscall.MS_SYNC),
		)
	}()

	outputSlice, freeOutputSlice, err := allocateSlice(int(*chunkSize * *chunkCount))
	if err != nil {
		panic(err)
	}
	defer freeOutputSlice()

	// Run the benchmark
	before := time.Now()

	copy(outputSlice, r3mappedSlice)

	after := time.Since(before)

	fmt.Printf("%.2f MB/s\n", float64(*chunkSize**chunkCount)/(1024*1024)/after.Seconds())

	// Validate the results
	if *check {
		remoteHash := xxhash.New()
		if _, err := io.Copy(remoteHash, bytes.NewReader(remoteSlice)); err != nil {
			panic(err)
		}

		localHash := xxhash.New()
		if _, err := io.Copy(localHash, bytes.NewReader(localSlice)); err != nil {
			panic(err)
		}

		r3mappedHash := xxhash.New()
		if _, err := io.Copy(r3mappedHash, bytes.NewReader(r3mappedSlice)); err != nil {
			panic(err)
		}

		outputHash := xxhash.New()
		if _, err := io.Copy(outputHash, bytes.NewReader(outputSlice)); err != nil {
			panic(err)
		}

		if remoteHash.Sum64() != localHash.Sum64() {
			panic("Remote, local, r3mapped and output hashes don't match")
		}

		fmt.Println("Remote, local, r3mapped and output hashes match.")
	}
}
