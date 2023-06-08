package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

func main() {
	laddr := flag.String("laddr", ":1337", "Listen address")

	size := flag.Int64("size", 4096*8192, "Size of the memory region to expose")

	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rb := backend.NewMemoryBackend(make([]byte, *size))

	b := lbackend.NewReaderAtBackend(
		chunks.NewChunkedReadWriterAt(
			rb,
			*chunkSize,
			*size / *chunkSize,
		),
		rb.Size,
		rb.Sync,
		false,
	)

	clients := 0

	registry := rpc.NewRegistry(
		services.NewBackend(b, *verbose),
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

	for {
		func() {
			conn, err := lis.Accept()
			if err != nil {
				log.Println("could not accept connection, continuing:", err)

				return
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
		}()
	}
}
