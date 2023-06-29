package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/mount/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
)

func main() {
	laddr := flag.String("addr", ":1337", "Listen address")
	size := flag.Int64("size", 4096*8192, "Size of the memory region to expose (ignored if file-based backend is used)")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	enableGrpc := flag.Bool("grpc", false, "Whether to use gRPC instead of Dudirekta")
	enableFrpc := flag.Bool("frpc", false, "Whether to use fRPC instead of Dudirekta")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	memory := flag.Bool("memory", false, "Whether to use memory instead of file-based backend")
	file := flag.String("file", "disk.img", "Path to file to expose (ignored if memory-based backend is used)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var svc *services.Backend
	if *memory {
		svc = services.NewBackend(backend.NewMemoryBackend(make([]byte, *size)), *verbose, *chunkSize)
	} else {
		file, err := os.OpenFile(*file, os.O_RDWR, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		svc = services.NewBackend(backend.NewFileBackend(file), *verbose, *chunkSize)
	}

	errs := make(chan error)
	if *enableGrpc {
		server := grpc.NewServer()

		v1proto.RegisterBackendServer(server, services.NewBackendGrpc(svc))

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
		server, err := v1frpc.NewServer(services.NewBackendFrpc(svc), nil, nil)
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

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
