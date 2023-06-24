package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/schollz/progressbar/v3"
)

var (
	errNoPeerFound = errors.New("no peer found")
)

func main() {
	raddr := flag.String("raddr", ":1337", "Listen address")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	registry := rpc.NewRegistry(
		&struct{}{},
		services.SeederRemote{},

		time.Second*10,
		ctx,
		&rpc.Options{
			ResponseBufferLen: rpc.DefaultResponseBufferLen,
			OnClientConnect: func(remoteID string) {
				ready <- struct{}{}
			},
		},
	)

	conn, err := net.Dial("tcp", *raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		if err := registry.Link(conn); err != nil {
			if !utils.IsClosedErr(err) {
				panic(err)
			}
		}
	}()

	<-ready

	log.Println("Connected to", conn.RemoteAddr())

	var peer *services.SeederRemote
	for _, candidate := range registry.Peers() {
		peer = &candidate

		break
	}

	if peer == nil {
		panic(errNoPeerFound)
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

		errs := make(chan error)
		go func() {
			if err := leecher.Wait(); err != nil {
				errs <- err

				return
			}

			close(errs)
		}()

		if err = leecher.Open(); err != nil {
			panic(err)
		}
		defer leecher.Close()

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		deviceSlice, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		log.Println("Connected to slice")

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

		errs := make(chan error)
		go func() {
			if err := leecher.Wait(); err != nil {
				errs <- err

				return
			}

			close(errs)
		}()

		if err = leecher.Open(); err != nil {
			panic(err)
		}
		defer leecher.Close()

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		deviceFile, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		log.Println("Connected on", deviceFile.Name())

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
