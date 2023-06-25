package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/schollz/progressbar/v3"
)

func main() {
	rawSize := flag.Int64("size", 4096*8192, "Size of the memory region to expose")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")
	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate in between Track() and Finalize()")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	seederErrs := make(chan error)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for err := range seederErrs {
			if err != nil {
				panic(err)
			}
		}
	}()

	var (
		svc               *services.Seeder
		invalidateLeecher func() error
	)
	if *slice {
		seeder := migration.NewSliceSeeder(
			backend.NewMemoryBackend(make([]byte, *rawSize)),

			&migration.SeederOptions{
				ChunkSize: *chunkSize,

				Verbose: *verbose,
			},

			nil,
			nil,
		)

		go func() {
			if err := seeder.Wait(); err != nil {
				seederErrs <- err

				return
			}

			close(seederErrs)
		}()

		deviceSlice, s, err := seeder.Open()
		if err != nil {
			panic(err)
		}
		defer seeder.Close()

		invalidateLeecher = func() error {
			copy(
				deviceSlice,
				make([]byte,
					int(math.Floor(
						float64(*rawSize)*(float64(*invalidate)/float64(100)),
					)),
				),
			)

			return nil
		}

		svc = s

		log.Println("Connected to slice")
	} else {
		seeder := migration.NewFileSeeder(
			backend.NewMemoryBackend(make([]byte, *rawSize)),

			&migration.SeederOptions{
				ChunkSize: *chunkSize,

				Verbose: *verbose,
			},
			&migration.FileSeederHooks{},

			nil,
			nil,
		)

		go func() {
			if err := seeder.Wait(); err != nil {
				seederErrs <- err

				return
			}

			close(seederErrs)
		}()

		deviceFile, s, err := seeder.Open()
		if err != nil {
			panic(err)
		}
		defer seeder.Close()

		invalidateLeecher = func() error {
			if _, err := io.CopyN(
				deviceFile,
				rand.Reader,
				int64(math.Floor(
					float64(*rawSize)*(float64(*invalidate)/float64(100)),
				)),
			); err != nil {
				return err
			}

			return nil
		}

		svc = s

		log.Println("Connected on", deviceFile.Name())
	}

	leecherErrs := make(chan error)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for err := range leecherErrs {
			if err != nil {
				panic(err)
			}
		}
	}()

	peer := &services.SeederRemote{
		ReadAt: svc.ReadAt,
		Size:   svc.Size,
		Track:  svc.Track,
		Sync:   svc.Sync,
		Close:  svc.Close,
	}

	size, err := peer.Size(ctx)
	if err != nil {
		panic(err)
	}

	bar := progressbar.NewOptions(
		int(size),
		progressbar.OptionSetDescription("Pulling"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionFullWidth(),
		// VT-100 compatibility
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	if *slice {
		leecher := migration.NewSliceLeecher(
			ctx,

			backend.NewMemoryBackend(make([]byte, size)),
			peer,

			&migration.LeecherOptions{
				ChunkSize: *chunkSize,

				PullWorkers: *pullWorkers,

				Verbose: *verbose,
			},
			&migration.SliceLeecherHooks{
				OnChunkIsLocal: func(off int64) error {
					bar.Add(int(*chunkSize))

					return nil
				},
				OnAfterSync: func(dirtyOffsets []int64) error {
					bar.Clear()

					delta := (len(dirtyOffsets) * int(*chunkSize))

					log.Printf("Invalidated: %.2f MB (%.2f Mb)", float64(delta)/(1024*1024), (float64(delta)/(1024*1024))*8)

					bar.ChangeMax(int(size) + delta)

					bar.Describe("Finalizing")

					return nil
				},
			},

			nil,
			nil,
		)

		go func() {
			if err := leecher.Wait(); err != nil {
				leecherErrs <- err

				return
			}

			close(leecherErrs)
		}()

		beforeOpen := time.Now()

		deviceSlice, err := leecher.Open()
		if err != nil {
			panic(err)
		}
		defer leecher.Close()

		afterOpen := time.Since(beforeOpen)

		fmt.Printf("Open: %v\n", afterOpen)

		if err := invalidateLeecher(); err != nil {
			panic(err)
		}

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		if err := leecher.Finalize(); err != nil {
			panic(err)
		}

		afterFinalize := time.Since(beforeFinalize)

		fmt.Printf("Finalize: %v\n", afterFinalize)

		output := make([]byte, size)

		beforeRead := time.Now()

		copy(output, deviceSlice)

		afterRead := time.Since(beforeRead)

		throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

		fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
	} else {
		leecher := migration.NewFileLeecher(
			ctx,

			backend.NewMemoryBackend(make([]byte, size)),
			peer,

			&migration.LeecherOptions{
				ChunkSize: *chunkSize,

				PullWorkers: *pullWorkers,

				Verbose: *verbose,
			},
			&migration.FileLeecherHooks{
				OnChunkIsLocal: func(off int64) error {
					bar.Add(int(*chunkSize))

					return nil
				},
				OnAfterSync: func(dirtyOffsets []int64) error {
					bar.Clear()

					delta := (len(dirtyOffsets) * int(*chunkSize))

					log.Printf("Invalidated: %.2f MB (%.2f Mb)", float64(delta)/(1024*1024), (float64(delta)/(1024*1024))*8)

					bar.ChangeMax(int(size) + delta)

					bar.Describe("Finalizing")

					return nil
				},
			},

			nil,
			nil,
		)

		go func() {
			if err := leecher.Wait(); err != nil {
				leecherErrs <- err

				return
			}

			close(leecherErrs)
		}()

		beforeOpen := time.Now()

		deviceFile, err := leecher.Open()
		if err != nil {
			panic(err)
		}
		defer leecher.Close()

		afterOpen := time.Since(beforeOpen)

		fmt.Printf("Open: %v\n", afterOpen)

		if err := invalidateLeecher(); err != nil {
			panic(err)
		}

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		if err := leecher.Finalize(); err != nil {
			panic(err)
		}

		afterFinalize := time.Since(beforeFinalize)

		fmt.Printf("Finalize: %v\n", afterFinalize)

		output := backend.NewMemoryBackend(make([]byte, size))

		beforeRead := time.Now()

		if _, err := io.CopyN(
			io.NewOffsetWriter(
				output,
				0,
			), deviceFile, size); err != nil {
			panic(err)
		}

		afterRead := time.Since(beforeRead)

		throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

		fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
	}
}
