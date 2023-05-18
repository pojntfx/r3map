package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/device"
)

func main() {
	size := flag.Int64("size", 4096*8192, "Size of the memory region to allocate")

	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")

	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	pullFirst := flag.Bool("pull-first", false, "Whether to completely pull from the remote before opening")

	pushWorkers := flag.Int64("push-workers", 512, "Push workers to launch in the background; pass in 0 to disable push")
	pushInterval := flag.Duration("push-interval", 5*time.Minute, "Interval after which to push changed chunks to the remote")

	check := flag.Bool("check", true, "Whether to check read and write results against expected data")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	remote := backend.NewMemoryBackend(make([]byte, *size), false)
	local := backend.NewMemoryBackend(make([]byte, *size), false)

	if _, err := io.CopyN(io.NewOffsetWriter(remote, 0), rand.Reader, *size); err != nil {
		panic(err)
	}

	mount := device.NewFileMount(
		ctx,

		remote,
		local,

		&device.MountOptions{
			ChunkSize: *chunkSize,

			PullWorkers: *pullWorkers,
			PullFirst:   *pullFirst,

			PushWorkers:  *pushWorkers,
			PushInterval: *pushInterval,

			Verbose: *verbose,
		},

		nil,
		nil,
	)

	go func() {
		if err := mount.Wait(); err != nil {
			panic(err)
		}
	}()

	beforeOpen := time.Now()

	mountedFile, err := mount.Open()
	if err != nil {
		panic(err)
	}
	defer func() {
		beforeClose := time.Now()

		_ = mount.Close()

		afterClose := time.Since(beforeClose)

		fmt.Printf("Close: %v\n", afterClose)
	}()

	afterOpen := time.Since(beforeOpen)

	fmt.Printf("Open: %v\n", afterOpen)

	beforeRead := time.Now()

	output := backend.NewMemoryBackend(make([]byte, *size), false)
	if _, err := io.Copy(io.NewOffsetWriter(output, 0), mountedFile); err != nil {
		panic(err)
	}

	afterRead := time.Since(beforeRead)

	fmt.Printf("Read throughput: %.2f MB/s\n", float64(*size)/(1024*1024)/afterRead.Seconds())

	validate := func(output io.ReaderAt) error {
		remoteHash := xxhash.New()
		if _, err := io.Copy(remoteHash, io.NewSectionReader(remote, 0, *size)); err != nil {
			return err
		}

		localHash := xxhash.New()
		if _, err := io.Copy(localHash, io.NewSectionReader(local, 0, *size)); err != nil {
			return err
		}

		mountedHash := xxhash.New()
		if _, err := io.Copy(mountedHash, io.NewSectionReader(mountedFile, 0, *size)); err != nil {
			return err
		}

		outputHash := xxhash.New()
		if _, err := io.Copy(outputHash, io.NewSectionReader(output, 0, *size)); err != nil {
			return err
		}

		if remoteHash.Sum64() != localHash.Sum64() {
			return errors.New("remote, local, mounted and output hashes don't match")
		}

		return nil
	}

	if *check {
		if err := validate(output); err != nil {
			panic(err)
		}

		fmt.Println("Read check: Passed")
	}

	beforeWrite := time.Now()

	if _, err := mountedFile.Seek(0, io.SeekStart); err != nil {
		panic(err)
	}

	if _, err := io.CopyN(mountedFile, rand.Reader, *size); err != nil {
		panic(err)
	}

	afterWrite := time.Since(beforeWrite)

	fmt.Printf("Write throughput: %.2f MB/s\n", float64(*size)/(1024*1024)/afterWrite.Seconds())

	if *check && *pushWorkers > 0 {
		beforeSync := time.Now()

		if err := mount.Sync(); err != nil {
			panic(err)
		}

		afterSync := time.Since(beforeSync)

		fmt.Printf("Sync: %v\n", afterSync)

		if err := validate(output); err != nil {
			panic(err)
		}

		fmt.Println("Write check: Passed")
	}
}
