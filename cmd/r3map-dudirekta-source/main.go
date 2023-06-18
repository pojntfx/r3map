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
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
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

	devicePath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		panic(err)
	}

	serverFile, err := os.Open(devicePath)
	if err != nil {
		panic(err)
	}
	defer serverFile.Close()

	rb := backend.NewMemoryBackend(make([]byte, *size))

	tr := chunks.NewTrackingReadWriterAt(rb)

	b := lbackend.NewReaderAtBackend(
		chunks.NewArbitraryReadWriterAt(
			chunks.NewChunkedReadWriterAt(
				tr,
				*chunkSize,
				*size / *chunkSize,
			),
			*chunkSize,
		),
		rb.Size,
		rb.Sync,
		false,
	)

	dev := device.NewDevice(
		b,
		serverFile,

		nil,
		nil,
	)
	defer dev.Close()

	errs := make(chan error)
	go func() {
		if err := dev.Wait(); err != nil {
			errs <- err

			return
		}
	}()

	if err := dev.Open(); err != nil {
		panic(err)
	}

	clients := 0

	registry := rpc.NewRegistry(
		services.NewSource(
			b,
			*verbose,
			func() ([]int64, error) {
				return tr.Flush(), nil
			},
		),
		struct{}{},

		time.Second*10,
		ctx,
		&rpc.Options{
			ResponseBufferLen: rpc.DefaultResponseBufferLen,
			OnClientConnect: func(remoteID string) {
				clients++

				log.Printf("%v clients connected", clients)

				tr.Track()
			},
			OnClientDisconnect: func(remoteID string) {
				clients--

				log.Printf("%v clients connected", clients)

				// TODO: If a remote called `Flush()` before, shut down
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
				log.Println("could not accept connection, continuing:", err)

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

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
