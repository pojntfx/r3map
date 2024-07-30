package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
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
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/panrpc/go/pkg/rpc"
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

	seederTypePanrpc = "panrpc"
	seederTypeGrpc   = "grpc"
	seederTypeFrpc   = "frpc"
)

var (
	knownBackendTypes = []string{backendTypeFile, backendTypeMemory, backendTypeDirectory}
	knownSeederTypes  = []string{seederTypePanrpc, seederTypeGrpc, seederTypeFrpc}

	errUnknownBackend = errors.New("unknown backend")
	errUnknownSeeder  = errors.New("unknown seeder")
)

func main() {
	s := flag.Int64("size", 536870912, "Size of the memory region, file to allocate or to size assume in case of the panrpc/gRPC/fRPC remotes")
	chunkSize := flag.Int64("chunk-size", client.MaximumBlockSize, "Chunk size to use")

	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background")

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
	seederLocation := flag.String("seeder-location", filepath.Join(os.TempDir(), "remote"), "Remote seeder's remote address (for panrpc/gRPC/fRPC, e.g. localhost:1337)")

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
		peer             *services.SeederRemote
		invalidateSeeder func() error
		seederBackend    backend.Backend
	)
	switch *seeder {
	case seederTypePanrpc:
		conn, err := net.Dial("tcp", *seederLocation)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		ready := make(chan struct{})
		registry := rpc.NewRegistry[services.SeederRemote, json.RawMessage](
			&struct{}{},

			ctx,

			&rpc.RegistryHooks{
				OnClientConnect: func(remoteID string) {
					ready <- struct{}{}
				},
			},
		)

		go func() {
			encoder := json.NewEncoder(conn)
			decoder := json.NewDecoder(conn)

			if err := registry.LinkStream(
				func(v rpc.Message[json.RawMessage]) error {
					return encoder.Encode(v)
				},
				func(v *rpc.Message[json.RawMessage]) error {
					return decoder.Decode(v)
				},

				func(v any) (json.RawMessage, error) {
					b, err := json.Marshal(v)
					if err != nil {
						return nil, err
					}

					return json.RawMessage(b), nil
				},
				func(data json.RawMessage, v any) error {
					return json.Unmarshal([]byte(data), v)
				},

				&rpc.LinkHooks{},
			); err != nil {
				if !utils.IsClosedErr(err) {
					seederErrs <- err

					return
				}
			}

			close(seederErrs)
		}()

		<-ready

		_ = registry.ForRemotes(func(remoteID string, remote services.SeederRemote) error {
			peer = &remote

			return nil
		})

		if peer == nil {
			panic(errNoPeerFound)
		}

		log.Println("Leeching from", *seederLocation)

	case seederTypeGrpc:
		conn, err := grpc.Dial(*seederLocation, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = conn.Close()

			close(seederErrs)
		}()

		log.Println("Leeching from", *seederLocation)

		peer = services.NewSeederRemoteGrpc(v1proto.NewSeederClient(conn))

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

		log.Println("Leeching from", *seederLocation)

		peer = services.NewSeederRemoteFrpc(client)

	case "":
		seederBackend = remote

		var svc *services.SeederService
		if *slice {
			mgr := migration.NewSliceMigrator(
				ctx,

				seederBackend,

				&migration.MigratorOptions{
					ChunkSize: *chunkSize,

					Verbose: *verbose,
				},
				nil,

				nil,
				nil,
			)

			go func() {
				if err := mgr.Wait(); err != nil {
					seederErrs <- err

					return
				}

				close(seederErrs)
			}()

			defer mgr.Close()
			deviceSlice, service, err := mgr.Seed()
			if err != nil {
				panic(err)
			}

			invalidateSeeder = func() error {
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
			mgr := migration.NewFileMigrator(
				ctx,

				seederBackend,

				&migration.MigratorOptions{
					ChunkSize: *chunkSize,

					Verbose: *verbose,
				},
				nil,

				nil,
				nil,
			)

			go func() {
				if err := mgr.Wait(); err != nil {
					seederErrs <- err

					return
				}

				close(seederErrs)
			}()

			defer mgr.Close()
			deviceFile, service, err := mgr.Seed()
			if err != nil {
				panic(err)
			}

			invalidateSeeder = func() error {
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
			Track:  svc.Track,
			Sync:   svc.Sync,
			Close:  svc.Close,
		}

	default:
		panic(errUnknownSeeder)
	}

	size := *s

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
		leecher := migration.NewSliceMigrator(
			ctx,

			local,

			&migration.MigratorOptions{
				ChunkSize: *chunkSize,

				PullWorkers: *pullWorkers,

				Verbose: *verbose,
			},
			&migration.MigratorHooks{
				OnAfterSync: func(dirtyOffsets []int64) error {
					bar.Clear()

					delta := (len(dirtyOffsets) * client.MaximumBlockSize)

					log.Printf("Invalidated: %.2f MB (%.2f Mb)", float64(delta)/(1024*1024), (float64(delta)/(1024*1024))*8)

					bar.ChangeMax(int(size) + delta)

					bar.Describe("Finalizing")

					return nil
				},

				OnChunkIsLocal: func(off int64) error {
					bar.Add(client.MaximumBlockSize)

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
		finalize, err := leecher.Leech(peer)
		if err != nil {
			panic(err)
		}

		afterOpen := time.Since(beforeOpen)

		fmt.Printf("Open: %v\n", afterOpen)

		if invalidateSeeder != nil {
			if err := invalidateSeeder(); err != nil {
				panic(err)
			}
		}

		log.Println("Press <ENTER> to finalize migration")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		_, deviceSlice, err := finalize()
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
		leecher := migration.NewFileMigrator(
			ctx,

			local,

			&migration.MigratorOptions{
				ChunkSize: *chunkSize,

				PullWorkers: *pullWorkers,

				Verbose: *verbose,
			},
			&migration.MigratorHooks{
				OnAfterSync: func(dirtyOffsets []int64) error {
					bar.Clear()

					delta := (len(dirtyOffsets) * client.MaximumBlockSize)

					log.Printf("Invalidated: %.2f MB (%.2f Mb)", float64(delta)/(1024*1024), (float64(delta)/(1024*1024))*8)

					bar.ChangeMax(int(size) + delta)

					bar.Describe("Finalizing")

					return nil
				},

				OnChunkIsLocal: func(off int64) error {
					bar.Add(client.MaximumBlockSize)

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
		finalize, err := leecher.Leech(peer)
		if err != nil {
			panic(err)
		}

		afterOpen := time.Since(beforeOpen)

		fmt.Printf("Open: %v\n", afterOpen)

		if invalidateSeeder != nil {
			if err := invalidateSeeder(); err != nil {
				panic(err)
			}
		}

		log.Println("Press <ENTER> to finalize migration")

		bufio.NewScanner(os.Stdin).Scan()

		beforeFinalize := time.Now()

		_, deviceFile, err := finalize()
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
