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
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

func main() {
	laddr := flag.String("addr", ":1337", "Listen address")

	memory := flag.Bool("memory", false, "Whether to use memory instead of file-based backend")
	memorySize := flag.Int64("memory-size", 4096*8192, "Size of the memory region to expose (ignored if file-based backend is used)")

	file := flag.String("file", "disk.img", "Path to file to expose (ignored if memory-based backend is used)")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var b backend.Backend
	if *memory {
		b = backend.NewMemoryBackend(make([]byte, *memorySize))
	} else {
		file, err := os.OpenFile(*file, os.O_RDWR, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		b = backend.NewFileBackend(file)
	}

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
