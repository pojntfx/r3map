package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/r3map/pkg/frontend"
)

const (
	backendTypeMemory = "memory"
	backendTypeFile   = "file"
)

var (
	errUnknownBackend = errors.New("unknown backend")
)

func main() {
	s := flag.Int64("size", 4096*8192, "Size of the memory region or file to allocate (ignored if a dudirekta remote is used)")

	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")

	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	pullFirst := flag.Bool("pull-first", false, "Whether to completely pull from the remote before opening")

	pushWorkers := flag.Int64("push-workers", 512, "Push workers to launch in the background; pass in 0 to disable push")
	pushInterval := flag.Duration("push-interval", 5*time.Minute, "Interval after which to push changed chunks to the remote")

	localBackend := flag.String(
		"local-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Local backend to use (one of %v)",
			[]string{
				backendTypeMemory,
				backendTypeFile,
			},
		),
	)
	remoteBackend := flag.String(
		"remote-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Remote backend to use (one of %v)",
			[]string{
				backendTypeMemory,
				backendTypeFile,
			},
		),
	)
	outputBackend := flag.String(
		"output-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Output backend to use (one of %v)",
			[]string{
				backendTypeMemory,
				backendTypeFile,
			},
		),
	)
	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")

	check := flag.Bool("check", true, "Whether to check read and write results against expected data")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		local  backend.Backend
		remote backend.Backend
		output backend.Backend
	)

	for _, config := range []struct {
		backendInstance *backend.Backend
		backendType     string
	}{
		{
			&local,
			*localBackend,
		},
		{
			&remote,
			*remoteBackend,
		},
		{
			&output,
			*outputBackend,
		},
	} {
		switch config.backendType {
		case backendTypeMemory:
			*config.backendInstance = backend.NewMemoryBackend(make([]byte, *s))

		case backendTypeFile:
			file, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			defer os.RemoveAll(file.Name())

			if err := file.Truncate(*s); err != nil {
				panic(err)
			}

			*config.backendInstance = backend.NewFileBackend(file)

		default:
			panic(errUnknownBackend)
		}
	}

	size, err := remote.Size()
	if err != nil {
		panic(err)
	}
	size = (size / *chunkSize) * *chunkSize

	if _, err := io.CopyN(io.NewOffsetWriter(remote, 0), rand.Reader, size); err != nil {
		panic(err)
	}

	beforeOpen := time.Now()

	var (
		mountedReader io.Reader
		mount         any
		sync          func() error
	)
	if *slice {
		mnt := frontend.NewSliceFrontend(
			ctx,

			remote,
			local,

			&frontend.Options{
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
			if err := mnt.Wait(); err != nil {
				panic(err)
			}
		}()

		mountedSlice, err := mnt.Open()
		if err != nil {
			panic(err)
		}

		defer func() {
			beforeClose := time.Now()

			_ = mnt.Close()

			afterClose := time.Since(beforeClose)

			fmt.Printf("Close: %v\n", afterClose)
		}()

		mountedReader, mount, sync = bytes.NewReader(mountedSlice), mountedSlice, mnt.Sync
	} else {
		mnt := frontend.NewFileFrontend(
			ctx,

			remote,
			local,

			&frontend.Options{
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
			if err := mnt.Wait(); err != nil {
				panic(err)
			}
		}()

		mountedFile, err := mnt.Open()
		if err != nil {
			panic(err)
		}

		defer func() {
			beforeClose := time.Now()

			_ = mnt.Close()

			afterClose := time.Since(beforeClose)

			fmt.Printf("Close: %v\n", afterClose)
		}()

		mountedReader, mount, sync = mountedFile, mountedFile, mnt.Sync
	}

	afterOpen := time.Since(beforeOpen)

	fmt.Printf("Open: %v\n", afterOpen)

	beforeRead := time.Now()

	if _, err := io.Copy(io.NewOffsetWriter(output, 0), mountedReader); err != nil {
		panic(err)
	}

	afterRead := time.Since(beforeRead)

	fmt.Printf("Read throughput: %.2f MB/s\n", float64(size)/(1024*1024)/afterRead.Seconds())

	validate := func(output io.ReaderAt) error {
		remoteHash := xxhash.New()
		if _, err := io.Copy(remoteHash, io.NewSectionReader(remote, 0, size)); err != nil {
			return err
		}

		localHash := xxhash.New()
		if _, err := io.Copy(localHash, io.NewSectionReader(local, 0, size)); err != nil {
			return err
		}

		mountedHash := xxhash.New()
		if _, err := io.Copy(mountedHash, mountedReader); err != nil {
			return err
		}

		outputHash := xxhash.New()
		if _, err := io.Copy(outputHash, io.NewSectionReader(output, 0, size)); err != nil {
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

	if !*slice {
		if _, err := mount.(*os.File).Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
	}

	beforeWrite := time.Now()

	if *slice {
		if _, err := rand.Read(mount.([]byte)); err != nil {
			panic(err)
		}
	} else {
		if _, err := io.CopyN(mount.(*os.File), rand.Reader, size); err != nil {
			panic(err)
		}
	}

	afterWrite := time.Since(beforeWrite)

	fmt.Printf("Write throughput: %.2f MB/s\n", float64(size)/(1024*1024)/afterWrite.Seconds())

	if *check && *pushWorkers > 0 {
		beforeSync := time.Now()

		if err := sync(); err != nil {
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
