package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	v1 "github.com/pojntfx/r3map/pkg/api/proto/v1"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	enableGRPC := flag.Bool("grpc", false, "Whether to use gRPC instead of Dudirekta")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var peer *services.SeederRemote
	if *enableGRPC {
		conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		log.Println("Connected to", *raddr)

		client := v1.NewSeederClient(conn)

		peer = &services.SeederRemote{
			ReadAt: func(ctx context.Context, length int, off int64) (r services.ReadAtResponse, err error) {
				res, err := client.ReadAt(ctx, &v1.ReadAtArgs{
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
				res, err := client.Size(ctx, &v1.SizeArgs{})
				if err != nil {
					return -1, err
				}

				return res.GetN(), nil
			},
			Track: func(ctx context.Context) error {
				if _, err := client.Track(ctx, &v1.TrackArgs{}); err != nil {
					return err
				}

				return nil
			},
			Sync: func(ctx context.Context) ([]int64, error) {
				res, err := client.Sync(ctx, &v1.SyncArgs{})
				if err != nil {
					return []int64{}, err
				}

				return res.GetDirtyOffsets(), nil
			},
			Close: func(ctx context.Context) error {
				if _, err := client.Close(ctx, &v1.CloseArgs{}); err != nil {
					return err
				}

				return nil
			},
		}
	} else {
		conn, err := net.Dial("tcp", *raddr)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

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

		go func() {
			if err := registry.Link(conn); err != nil {
				if !utils.IsClosedErr(err) {
					panic(err)
				}
			}
		}()

		<-ready

		log.Println("Connected to", conn.RemoteAddr())

		for _, candidate := range registry.Peers() {
			peer = &candidate

			break
		}

		if peer == nil {
			panic(errNoPeerFound)
		}
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

	bar.Add64(*chunkSize)

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

		errs := make(chan error)
		go func() {
			if err := leecher.Wait(); err != nil {
				errs <- err

				return
			}

			close(errs)
		}()

		if err := leecher.Open(); err != nil {
			panic(err)
		}
		defer leecher.Close()

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		deviceSlice, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		bar.Clear()

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

		errs := make(chan error)
		go func() {
			if err := leecher.Wait(); err != nil {
				errs <- err

				return
			}

			close(errs)
		}()

		if err := leecher.Open(); err != nil {
			panic(err)
		}
		defer leecher.Close()

		bar.Clear()

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		deviceFile, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		bar.Clear()

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
