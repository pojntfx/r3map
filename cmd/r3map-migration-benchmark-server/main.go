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
	"net"
	"os"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/migration/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
)

func main() {
	laddr := flag.String("laddr", ":1337", "Listen address")
	size := flag.Int64("size", 4096*8192, "Size of the memory region to expose")
	slice := flag.Bool("slice", false, "Whether to use the slice frontend instead of the file frontend")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	enableGrpc := flag.Bool("grpc", false, "Whether to use gRPC instead of Dudirekta")
	enableFrpc := flag.Bool("frpc", false, "Whether to use fRPC instead of Dudirekta")
	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate in between Track() and Finalize()")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := backend.NewMemoryBackend(make([]byte, *size))

	var (
		svc               *services.Seeder
		invalidateLeecher func() error
		errs              = make(chan error)
	)
	if *slice {
		seeder := migration.NewSliceSeeder(
			b,

			&migration.SeederOptions{
				ChunkSize: *chunkSize,

				Verbose: *verbose,
			},

			nil,
			nil,
		)

		go func() {
			if err := seeder.Wait(); err != nil {
				errs <- err

				return
			}

			close(errs)
		}()

		defer seeder.Close()
		deviceSlice, s, err := seeder.Open()
		if err != nil {
			panic(err)
		}

		invalidateLeecher = func() error {
			if _, err := io.CopyN(
				utils.NewSliceWriter(deviceSlice),
				rand.Reader,
				int64(math.Floor(
					float64(*size)*(float64(*invalidate)/float64(100)),
				)),
			); err != nil {
				return err
			}

			return nil
		}

		svc = s

		log.Println("Connected to slice")
	} else {
		seeder := migration.NewFileSeeder(
			b,

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
				errs <- err

				return
			}

			close(errs)
		}()

		defer seeder.Close()
		deviceFile, s, err := seeder.Open()
		if err != nil {
			panic(err)
		}

		invalidateLeecher = func() error {
			if _, err := io.CopyN(
				deviceFile,
				rand.Reader,
				int64(math.Floor(
					float64(*size)*(float64(*invalidate)/float64(100)),
				)),
			); err != nil {
				return err
			}

			return nil
		}

		svc = s

		log.Println("Connected on", deviceFile.Name())
	}

	if *enableGrpc {
		server := grpc.NewServer()

		v1proto.RegisterSeederServer(server, services.NewSeederGrpc(svc))

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
		server, err := v1frpc.NewServer(services.NewSeederFrpc(svc), nil, nil)
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
		registry := rpc.NewRegistry(
			svc,
			struct{}{},

			time.Second*10,
			ctx,
			&rpc.Options{
				ResponseBufferLen: rpc.DefaultResponseBufferLen,
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

					if err := registry.Link(conn); err != nil {
						panic(err)
					}
				}()
			}
		}()
	}

	go func() {
		log.Println("Press <ENTER> to invalidate")

		bufio.NewScanner(os.Stdin).Scan()

		beforeInvalidate := time.Now()

		if err := invalidateLeecher(); err != nil {
			errs <- err

			return
		}

		afterInvalidate := time.Since(beforeInvalidate)

		fmt.Printf("Invalidate: %v\n", afterInvalidate)
	}()

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
