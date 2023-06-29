package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/migration/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
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
	s := flag.Int64("size", 4096*8192, "Size of the memory region to expose. Will be ignored if a remote seeder is used.")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")
	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate in between Track() and Finalize(). Will be ignored if a remote seeder is used.")
	check := flag.Bool("check", true, "Whether to check read results against expected data")
	raddr := flag.String("raddr", "", "Listen address for remote seeder (leave empty to create a local seeder)")
	enableGrpc := flag.Bool("grpc", false, "Whether to use gRPC instead of Dudirekta for the remote seeder")
	enableFrpc := flag.Bool("frpc", false, "Whether to use fRPC instead of Dudirekta for the remote seeder")

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
		peer              *services.SeederRemote
		invalidateLeecher func() error
		seederBackend     backend.Backend
	)
	if strings.TrimSpace(*raddr) == "" {
		var svc *services.Seeder
		if *slice {
			seederBackend = backend.NewMemoryBackend(make([]byte, *s))

			seeder := migration.NewSliceSeeder(
				seederBackend,

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

			defer seeder.Close()
			deviceSlice, service, err := seeder.Open()
			if err != nil {
				panic(err)
			}

			invalidateLeecher = func() error {
				if _, err := io.CopyN(
					utils.NewSliceWriter(deviceSlice),
					rand.Reader,
					int64(math.Floor(
						float64(*s)*(float64(*invalidate)/float64(100)),
					)),
				); err != nil {
					return err
				}

				return nil
			}

			svc = service

			log.Println("Connected to slice")
		} else {
			seederBackend = backend.NewMemoryBackend(make([]byte, *s))

			seeder := migration.NewFileSeeder(
				seederBackend,

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

			defer seeder.Close()
			deviceFile, service, err := seeder.Open()
			if err != nil {
				panic(err)
			}

			invalidateLeecher = func() error {
				if _, err := io.CopyN(
					deviceFile,
					rand.Reader,
					int64(math.Floor(
						float64(*s)*(float64(*invalidate)/float64(100)),
					)),
				); err != nil {
					return err
				}

				return nil
			}

			svc = service

			log.Println("Connected on", deviceFile.Name())
		}

		peer = &services.SeederRemote{
			ReadAt: svc.ReadAt,
			Size:   svc.Size,
			Track:  svc.Track,
			Sync:   svc.Sync,
			Close:  svc.Close,
		}
	} else {
		if *enableGrpc {
			conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				panic(err)
			}
			defer func() {
				_ = conn.Close()

				close(seederErrs)
			}()

			log.Println("Connected to", *raddr)

			client := v1proto.NewSeederClient(conn)

			peer = &services.SeederRemote{
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
		} else if *enableFrpc {
			client, err := v1frpc.NewClient(nil, nil)
			if err != nil {
				panic(err)
			}

			if err := client.Connect(*raddr); err != nil {
				panic(err)
			}
			defer func() {
				_ = client.Close()

				close(seederErrs)
			}()

			log.Println("Connected to", *raddr)

			peer = &services.SeederRemote{
				ReadAt: func(ctx context.Context, length int, off int64) (r services.ReadAtResponse, err error) {
					res, err := client.Seeder.ReadAt(ctx, &v1frpc.ComPojtingerFelicitasR3MapMigrationV1ReadAtArgs{
						Length: int32(length),
						Off:    off,
					})
					if err != nil {
						return services.ReadAtResponse{}, err
					}

					return services.ReadAtResponse{
						N: int(res.N),
						P: res.P,
					}, err
				},
				Size: func(ctx context.Context) (int64, error) {
					res, err := client.Seeder.Size(ctx, &v1frpc.ComPojtingerFelicitasR3MapMigrationV1SizeArgs{})
					if err != nil {
						return -1, err
					}

					return res.N, nil
				},
				Track: func(ctx context.Context) error {
					if _, err := client.Seeder.Track(ctx, &v1frpc.ComPojtingerFelicitasR3MapMigrationV1TrackArgs{}); err != nil {
						return err
					}

					return nil
				},
				Sync: func(ctx context.Context) ([]int64, error) {
					res, err := client.Seeder.Sync(ctx, &v1frpc.ComPojtingerFelicitasR3MapMigrationV1SyncArgs{})
					if err != nil {
						return []int64{}, err
					}

					return res.DirtyOffsets, nil
				},
				Close: func(ctx context.Context) error {
					if _, err := client.Seeder.Close(ctx, &v1frpc.ComPojtingerFelicitasR3MapMigrationV1CloseArgs{}); err != nil {
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
						seederErrs <- err

						return
					}
				}

				close(seederErrs)
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

	var (
		leecherBackend backend.Backend
		outputReader   io.Reader
	)
	if *slice {
		leecherBackend = backend.NewMemoryBackend(make([]byte, size))

		leecher := migration.NewSliceLeecher(
			ctx,

			leecherBackend,
			peer,

			&migration.LeecherOptions{
				ChunkSize: *chunkSize,

				PullWorkers: *pullWorkers,

				Verbose: *verbose,
			},
			&migration.SliceLeecherHooks{
				OnChunkIsLocal: func(off int64) error {
					bar.Add64(*chunkSize)

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

		defer leecher.Close()
		if err := leecher.Open(); err != nil {
			panic(err)
		}

		afterOpen := time.Since(beforeOpen)

		fmt.Printf("Open: %v\n", afterOpen)

		if invalidateLeecher != nil {
			if err := invalidateLeecher(); err != nil {
				panic(err)
			}
		}

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		deviceSlice, err := leecher.Finalize()
		if err != nil {
			panic(err)
		}

		afterFinalize := time.Since(beforeFinalize)

		bar.Clear()

		fmt.Printf("Finalize: %v\n", afterFinalize)

		output := make([]byte, size)

		beforeRead := time.Now()

		copy(output, deviceSlice)

		afterRead := time.Since(beforeRead)

		throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

		fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)

		outputReader = bytes.NewReader(output)
	} else {
		leecherBackend = backend.NewMemoryBackend(make([]byte, size))

		leecher := migration.NewFileLeecher(
			ctx,

			leecherBackend,
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

		defer leecher.Close()
		if err := leecher.Open(); err != nil {
			panic(err)
		}

		afterOpen := time.Since(beforeOpen)

		fmt.Printf("Open: %v\n", afterOpen)

		if invalidateLeecher != nil {
			if err := invalidateLeecher(); err != nil {
				panic(err)
			}
		}

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

		output := make([]byte, size)

		beforeRead := time.Now()

		if _, err := io.CopyN(
			io.NewOffsetWriter(
				backend.NewMemoryBackend(output),
				0,
			), deviceFile, size); err != nil {
			panic(err)
		}

		afterRead := time.Since(beforeRead)

		throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

		fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)

		outputReader = bytes.NewReader(output)
	}

	if *check {
		seederHash := xxhash.New()
		if seederBackend != nil {
			if _, err := io.Copy(
				seederHash,
				io.NewSectionReader(
					seederBackend,
					0,
					size,
				),
			); err != nil {
				panic(err)
			}
		}

		leecherHash := xxhash.New()
		if _, err := io.Copy(
			leecherHash,
			io.NewSectionReader(
				leecherBackend,
				0,
				size,
			),
		); err != nil {
			panic(err)
		}

		outputHash := xxhash.New()
		if _, err := io.CopyN(
			outputHash,
			outputReader,
			size,
		); err != nil {
			panic(err)
		}

		if seederBackend == nil {
			if leecherHash.Sum64() != outputHash.Sum64() {
				panic("local, and output hashes don't match")
			}
		} else {
			if !(seederHash.Sum64() == leecherHash.Sum64() && leecherHash.Sum64() == outputHash.Sum64()) {
				panic("remote, local, and output hashes don't match")
			}
		}

		fmt.Println("Read check: Passed")
	}
}
