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
	"path/filepath"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/migration/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
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

const (
	backendTypeFile      = "file"
	backendTypeMemory    = "memory"
	backendTypeDirectory = "directory"

	seederTypeDudirekta = "dudirekta"
	seederTypeGrpc      = "grpc"
	seederTypeFrpc      = "frpc"
)

var (
	knownBackendTypes = []string{backendTypeFile, backendTypeMemory, backendTypeDirectory}
	knownSeederTypes  = []string{seederTypeDudirekta, seederTypeGrpc, seederTypeFrpc}

	errUnknownBackend = errors.New("unknown backend")
	errUnknownSeeder  = errors.New("unknown seeder")
)

func main() {
	s := flag.Int64("size", 4096*8192, "Size of the memory region, file to allocate or to size assume in case of the dudirekta/gRPC/fRPC remotes")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")

	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")

	localBackend := flag.String(
		"local-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Local backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	localLocation := flag.String("local-backend-location", filepath.Join(os.TempDir(), "local"), "Local backend's directory (for directory backend)")
	localChunking := flag.Bool("local-backend-chunking", false, "Whether the local backend requires to be interfaced with in fixed chunks")

	remoteBackend := flag.String(
		"remote-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Remote backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	remoteLocation := flag.String("remote-backend-location", filepath.Join(os.TempDir(), "remote"), "Remote backend's directory (for directory backend)")
	remoteChunking := flag.Bool("remote-backend-chunking", false, "Whether the remote backend requires to be interfaced with in fixed chunks")

	outputBackend := flag.String(
		"output-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Output backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	outputLocation := flag.String("output-backend-location", filepath.Join(os.TempDir(), "output"), "Output backend's or directory (for directory backend)")
	outputChunking := flag.Bool("output-backend-chunking", false, "Whether the output backend requires to be interfaced with in fixed chunks")

	seeder := flag.String(
		"seeder",
		"",
		fmt.Sprintf(
			"Remote seeder to use (one of %v) (keep empty to use remote backend instead)",
			knownSeederTypes,
		),
	)
	seederLocation := flag.String("seeder-location", filepath.Join(os.TempDir(), "remote"), "Remote seeder's remote address (for dudirekta/gRPC/fRPC, e.g. localhost:1337)")

	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")
	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate in between Track() and Finalize(). Will be ignored if a remote seeder is used.")

	check := flag.Bool("check", true, "Whether to check read results against expected data")
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
		backendLocation string
		chunking        bool
	}{
		{
			&local,
			*localBackend,
			*localLocation,
			*localChunking,
		},
		{
			&remote,
			*remoteBackend,
			*remoteLocation,
			*remoteChunking,
		},
		{
			&output,
			*outputBackend,
			*outputLocation,
			*outputChunking,
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

		case backendTypeDirectory:
			if err := os.MkdirAll(config.backendLocation, os.ModePerm); err != nil {
				panic(err)
			}

			*config.backendInstance = lbackend.NewDirectoryBackend(config.backendLocation, *s, *chunkSize, 512, false)

		default:
			panic(errUnknownBackend)
		}

		if config.chunking {
			*config.backendInstance = lbackend.NewReaderAtBackend(
				chunks.NewArbitraryReadWriterAt(
					*config.backendInstance,
					*chunkSize,
				),
				(*config.backendInstance).Size,
				(*config.backendInstance).Sync,
				false,
			)
		}
	}

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
	switch *seeder {
	case seederTypeDudirekta:
		conn, err := net.Dial("tcp", *seederLocation)
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

		for _, candidate := range registry.Peers() {
			peer = &candidate

			break
		}

		if peer == nil {
			panic(errNoPeerFound)
		}

	case seederTypeGrpc:
		conn, err := grpc.Dial(*seederLocation, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = conn.Close()

			close(seederErrs)
		}()

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

	case seederTypeFrpc:
		client, err := v1frpc.NewClient(nil, nil)
		if err != nil {
			panic(err)
		}

		if err := client.Connect(*seederLocation); err != nil {
			panic(err)
		}
		defer func() {
			_ = client.Close()

			close(seederErrs)
		}()

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

	case "":
		seederBackend = remote

		var svc *services.Seeder
		if *slice {

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
		} else {
			seeder := migration.NewFileSeeder(
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
		}

		peer = &services.SeederRemote{
			ReadAt: svc.ReadAt,
			Size:   svc.Size,
			Track:  svc.Track,
			Sync:   svc.Sync,
			Close:  svc.Close,
		}

	default:
		panic(errUnknownSeeder)
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

	if *slice {
		leecher := migration.NewSliceLeecher(
			ctx,

			local,
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

		beforeRead := time.Now()

		if _, err := io.CopyN(
			io.NewOffsetWriter(
				output,
				0,
			), bytes.NewReader(deviceSlice), size); err != nil {
			panic(err)
		}

		afterRead := time.Since(beforeRead)

		throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

		fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
	} else {
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
				local,
				0,
				size,
			),
		); err != nil {
			panic(err)
		}

		outputHash := xxhash.New()
		if _, err := io.CopyN(
			outputHash,
			io.NewSectionReader(
				output,
				0,
				size,
			),
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
