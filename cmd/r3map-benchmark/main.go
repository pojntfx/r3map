package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/frontend"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/redis/go-redis/v9"
)

const (
	backendTypeFile      = "file"
	backendTypeMemory    = "memory"
	backendTypeDirectory = "directory"
	backendTypeDudirekta = "dudirekta"
	backendTypeRedis     = "redis"
)

var (
	knownBackendTypes = []string{backendTypeFile, backendTypeMemory, backendTypeDirectory, backendTypeDudirekta, backendTypeRedis}

	errUnknownBackend = errors.New("unknown backend")
	errNoPeerFound    = errors.New("no peer found")
)

func main() {
	s := flag.Int64("size", 4096*8192, "Size of the memory region, file to allocate or to size assume in case of the dudirekta remote")

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
			knownBackendTypes,
		),
	)
	localLocation := flag.String("local-location", filepath.Join(os.TempDir(), "local"), "Local backend's remote address (for dudirekta, e.g. localhost:1337), URI (for redis, e.g. redis://username:password@localhost:6379/0) or directory (for directory backend)")
	localChunking := flag.Bool("local-chunking", true, "Whether the local backend requires to be interfaced with in fixed chunks in tests")

	remoteBackend := flag.String(
		"remote-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Remote backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	remoteLocation := flag.String("remote-location", filepath.Join(os.TempDir(), "remote"), "Remote backend's remote address (for dudirekta, e.g. localhost:1337), URI (for redis, e.g. redis://username:password@localhost:6379/0) or directory (for directory backend)")
	remoteChunking := flag.Bool("remote-chunking", true, "Whether the remote backend requires to be interfaced with in fixed chunks in tests")

	outputBackend := flag.String(
		"output-backend",
		backendTypeFile,
		fmt.Sprintf(
			"Output backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	outputLocation := flag.String("output-location", filepath.Join(os.TempDir(), "output"), "Output backend's output address (for dudirekta, e.g. localhost:1337), URI (for redis, e.g. redis://username:password@localhost:6379/0) or directory (for directory backend)")
	outputChunking := flag.Bool("output-chunking", false, "Whether the output backend requires to be interfaced with in fixed chunks in tests")

	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")

	check := flag.Bool("check", true, "Whether to check read and write results against expected data")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		local backend.Backend
		l     backend.Backend

		remote backend.Backend
		r      backend.Backend

		output backend.Backend
	)

	for _, config := range []struct {
		backendInstance              *backend.Backend
		backendTestInstance          *backend.Backend
		backendType                  string
		backendLocation              string
		testInstanceRequiresChunking bool
	}{
		{
			&l,
			&local,
			*localBackend,
			*localLocation,
			*localChunking,
		},
		{
			&r,
			&remote,
			*remoteBackend,
			*remoteLocation,
			*remoteChunking,
		},
		{
			&output,
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

		case backendTypeDudirekta:
			ready := make(chan struct{})
			registry := rpc.NewRegistry(
				&struct{}{},
				services.BackendRemote{},

				time.Second*10,
				ctx,
				&rpc.Options{
					ResponseBufferLen: rpc.DefaultResponseBufferLen,
					OnClientConnect: func(remoteID string) {
						ready <- struct{}{}
					},
				},
			)

			conn, err := net.Dial("tcp", config.backendLocation)
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

			var peer *services.BackendRemote
			for _, candidate := range registry.Peers() {
				peer = &candidate

				break
			}

			if peer == nil {
				panic(errNoPeerFound)
			}

			*config.backendInstance = lbackend.NewRPCBackend(ctx, *peer, false)

		case backendTypeRedis:
			u, err := redis.ParseURL(config.backendLocation)
			if err != nil {
				panic(err)
			}

			*config.backendInstance = lbackend.NewRedisBackend(ctx, u, *s, false)

		default:
			panic(errUnknownBackend)
		}

		if config.testInstanceRequiresChunking {
			*config.backendTestInstance = lbackend.NewReaderAtBackend(
				chunks.NewArbitraryReadWriterAt(
					chunks.NewChunkedReadWriterAt(
						*config.backendInstance, *chunkSize, *s / *chunkSize),
					*chunkSize,
				),
				(*config.backendInstance).Size,
				(*config.backendInstance).Sync,
				false,
			)
		} else {
			*config.backendTestInstance = *config.backendInstance
		}
	}

	size := (*s / *chunkSize) * *chunkSize

	if _, err := io.CopyN(
		io.NewOffsetWriter(
			remote,
			0), rand.Reader, size); err != nil {
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

			r,
			l,

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

			r,
			l,

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

	if _, err := io.CopyN(
		io.NewOffsetWriter(
			output,
			0,
		), mountedReader, size); err != nil {
		panic(err)
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)

	validate := func(output io.ReaderAt) error {
		if *slice {
			mountedReader = bytes.NewReader(mount.([]byte))
		} else {
			if _, err := mount.(*os.File).Seek(0, io.SeekStart); err != nil {
				panic(err)
			}
		}

		remoteHash := xxhash.New()
		if _, err := io.Copy(
			remoteHash,
			io.NewSectionReader(
				remote,
				0,
				size,
			),
		); err != nil {
			return err
		}

		localHash := xxhash.New()
		if _, err := io.Copy(
			localHash,
			io.NewSectionReader(
				local,
				0,
				size,
			),
		); err != nil {
			return err
		}

		outputHash := xxhash.New()
		if _, err := io.Copy(
			outputHash,
			io.NewSectionReader(
				output,
				0,
				size,
			),
		); err != nil {
			return err
		}

		if !(remoteHash.Sum64() == localHash.Sum64() && localHash.Sum64() == outputHash.Sum64()) {
			return errors.New("remote, local, and output hashes don't match")
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

	throughputMB = float64(size) / (1024 * 1024) / afterWrite.Seconds()

	fmt.Printf("Write throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)

	if *check && *pushWorkers > 0 {
		if *slice {
			mountedReader = bytes.NewReader(mount.([]byte))
		} else {
			if _, err := mount.(*os.File).Seek(0, io.SeekStart); err != nil {
				panic(err)
			}
		}

		if _, err := io.CopyN(
			io.NewOffsetWriter(
				output,
				0,
			), mountedReader, size); err != nil {
			panic(err)
		}

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
