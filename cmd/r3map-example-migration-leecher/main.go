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
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	size := flag.Int64("size", 671088640, "Size of the memory region, file to allocate or to size assume in case of the dudirekta/gRPC/fRPC remotes")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")

	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")

	raddr := flag.String("raddr", "localhost:1337", "Remote address")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	localFile, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(localFile.Name())

	if err := localFile.Truncate(*size); err != nil {
		panic(err)
	}

	local := backend.NewFileBackend(localFile)

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

	conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = conn.Close()

		close(seederErrs)
	}()

	client := v1proto.NewSeederClient(conn)

	peer := &services.SeederRemote{
		ReadAt: func(ctx context.Context, length int, off int64) (r services.ReadAtResponse, err error) {
			res, err := client.ReadAt(ctx, &v1proto.ReadAtArgs{
				Length: int32(length),
				Off:    off,
			})
			if err != nil {
				return services.ReadAtResponse{}, err
			}

			return services.ReadAtResponse{
				N: int(res.GetN()),
				P: res.GetP(),
			}, err
		},
		Size: func(ctx context.Context) (int64, error) {
			res, err := client.Size(ctx, &v1proto.SizeArgs{})
			if err != nil {
				return -1, err
			}

			return res.GetN(), nil
		},
		Track: func(ctx context.Context) error {
			if _, err := client.Track(ctx, &v1proto.TrackArgs{}); err != nil {
				return err
			}

			return nil
		},
		Sync: func(ctx context.Context) ([]int64, error) {
			res, err := client.Sync(ctx, &v1proto.SyncArgs{})
			if err != nil {
				return []int64{}, err
			}

			return res.GetDirtyOffsets(), nil
		},
		Close: func(ctx context.Context) error {
			if _, err := client.Close(ctx, &v1proto.CloseArgs{}); err != nil {
				return err
			}

			return nil
		},
	}

	bar := progressbar.NewOptions(
		int(*size),
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

	bar.Add64(*chunkSize)

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

	leecher := migration.NewFileLeecher(
		ctx,

		local,
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

				bar.ChangeMax(int(*size) + delta)

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

	defer leecher.Close()
	if err := leecher.Open(); err != nil {
		panic(err)
	}

	afterOpen := time.Since(beforeOpen)

	fmt.Printf("Open: %v\n", afterOpen)

	log.Println("Press <ENTER> to finalize")

	bufio.NewScanner(os.Stdin).Scan()

	beforeFinalize := time.Now()

	deviceFile, err := leecher.Finalize()
	if err != nil {
		panic(err)
	}

	afterFinalize := time.Since(beforeFinalize)

	bar.Clear()

	fmt.Printf("Finalize: %v\n", afterFinalize)

	beforeRead := time.Now()

	if _, err := io.CopyN(io.Discard, deviceFile, *size); err != nil {
		panic(err)
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(*size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
}
