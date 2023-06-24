package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
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

	var svc *services.Seeder
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

		_, s, err := seeder.Open()
		if err != nil {
			panic(err)
		}
		defer seeder.Close()

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

	chunkCount := int(size / *chunkSize)

	bar := progressbar.NewOptions(
		chunkCount,
		progressbar.OptionSetDescription("Pulling"),
		progressbar.OptionSetItsString("chunk"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
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
					bar.Add(1)

					return nil
				},
				OnAfterSync: func(dirtyOffsets []int64) error {
					bar.Clear()

					log.Printf("Invalidated %v dirty offsets", len(dirtyOffsets))

					bar.ChangeMax(chunkCount + len(dirtyOffsets))

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

		if err = leecher.Open(); err != nil {
			panic(err)
		}
		defer leecher.Close()

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		deviceSlice, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		afterFinalize := time.Since(beforeFinalize)

		fmt.Printf("Finalization latency: %v\n", afterFinalize)

		output := make([]byte, size)

		copy(output, deviceSlice)
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
					bar.Add(1)

					return nil
				},
				OnAfterSync: func(dirtyOffsets []int64) error {
					bar.Clear()

					log.Printf("Invalidated %v dirty offsets", len(dirtyOffsets))

					bar.ChangeMax(chunkCount + len(dirtyOffsets))

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

		if err = leecher.Open(); err != nil {
			panic(err)
		}
		defer leecher.Close()

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		deviceFile, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		afterFinalize := time.Since(beforeFinalize)

		fmt.Printf("Finalization latency: %v\n", afterFinalize)

		output := backend.NewMemoryBackend(make([]byte, size))

		if _, err := io.CopyN(
			io.NewOffsetWriter(
				output,
				0,
			), deviceFile, size); err != nil {
			panic(err)
		}
	}
}
