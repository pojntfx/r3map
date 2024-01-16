package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/panrpc/pkg/rpc"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/mount/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
)

const (
	backendTypeFile      = "file"
	backendTypeMemory    = "memory"
	backendTypeDirectory = "directory"
)

var (
	knownBackendTypes = []string{backendTypeFile, backendTypeMemory, backendTypeDirectory}

	errUnknownBackend = errors.New("unknown backend")
)

func main() {
	laddr := flag.String("addr", ":1337", "Listen address")
	enableGrpc := flag.Bool("grpc", false, "Whether to use gRPC instead of panrpc")
	enableFrpc := flag.Bool("frpc", false, "Whether to use fRPC instead of panrpc")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	size := flag.Int64("size", 536870912, "Size of the memory region or file to allocate")
	chunkSize := flag.Int64("chunk-size", client.MaximumBlockSize, "Chunk size to use")
	maxChunkSize := flag.Int64("max-chunk-size", services.MaxChunkSize, "Maximum chunk size to support")
	bck := flag.String(
		"backend",
		backendTypeFile,
		fmt.Sprintf(
			"Backend to use (one of %v)",
			knownBackendTypes,
		),
	)
	location := flag.String("location", filepath.Join(os.TempDir(), "local"), "Backend's directory (for directory backend)")
	chunking := flag.Bool("chunking", false, "Whether the backend requires to be interfaced with in fixed chunks")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var b backend.Backend
	switch *bck {
	case backendTypeMemory:
		b = backend.NewMemoryBackend(make([]byte, *size))

	case backendTypeFile:
		file, err := os.CreateTemp("", "")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(file.Name())

		if err := file.Truncate(*size); err != nil {
			panic(err)
		}

		b = backend.NewFileBackend(file)

	case backendTypeDirectory:
		if err := os.MkdirAll(*location, os.ModePerm); err != nil {
			panic(err)
		}

		b = lbackend.NewDirectoryBackend(*location, *size, *chunkSize, 512, false)
	default:
		panic(errUnknownBackend)
	}

	if *chunking {
		b = lbackend.NewReaderAtBackend(
			chunks.NewArbitraryReadWriterAt(
				b,
				*chunkSize,
			),
			(b).Size,
			(b).Sync,
			false,
		)
	}

	var (
		svc  = services.NewBackend(b, *verbose, *maxChunkSize)
		errs = make(chan error)
	)
	if *enableGrpc {
		server := grpc.NewServer()

		v1proto.RegisterBackendServer(server, services.NewBackendServiceGrpc(svc))

		lis, err := net.Listen("tcp", *laddr)
		if err != nil {
			panic(err)
		}
		defer lis.Close()

		log.Println("Listening on", lis.Addr())

		go func() {
			if err := server.Serve(lis); err != nil {
				if !utils.IsClosedErr(err) {
					errs <- err
				}

				return
			}
		}()
	} else if *enableFrpc {
		server, err := v1frpc.NewServer(services.NewBackendServiceFrpc(svc), nil, nil)
		if err != nil {
			panic(err)
		}

		log.Println("Listening on", *laddr)

		go func() {
			if err := server.Start(*laddr); err != nil {
				if !utils.IsClosedErr(err) {
					errs <- err
				}

				return
			}
		}()
	} else {
		clients := 0
		registry := rpc.NewRegistry[struct{}, json.RawMessage](
			svc,

			ctx,

			&rpc.Options{
				OnClientConnect: func(remoteID string) {
					clients++

					log.Printf("%v clients connected", clients)
				},
				OnClientDisconnect: func(remoteID string) {
					clients--

					log.Printf("%v clients connected", clients)
				},
			},
		)

		lis, err := net.Listen("tcp", *laddr)
		if err != nil {
			panic(err)
		}
		defer lis.Close()

		log.Println("Listening on", lis.Addr())

		go func() {
			for {
				conn, err := lis.Accept()
				if err != nil {
					if !utils.IsClosedErr(err) {
						log.Println("could not accept connection, continuing:", err)
					}

					continue
				}

				go func() {
					defer func() {
						_ = conn.Close()

						if err := recover(); err != nil {
							if !utils.IsClosedErr(err.(error)) {
								log.Printf("Client disconnected with error: %v", err)
							}
						}
					}()

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
					); err != nil {
						panic(err)
					}
				}()
			}
		}()
	}

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
