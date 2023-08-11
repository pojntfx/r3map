package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/pojntfx/go-nbd/pkg/backend"
	v1 "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
)

func main() {
	size := flag.Int64("size", 4096*8192*10, "Size of the resource")

	maxChunkSize := flag.Int64("max-chunk-size", services.MaxChunkSize, "Maximum chunk size to support")

	laddr := flag.String("laddr", "localhost:1337", "Listen address")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	f, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(f.Name())

	if err := f.Truncate(*size); err != nil {
		panic(err)
	}

	srv := grpc.NewServer()

	v1.RegisterBackendServer(
		srv,
		services.NewBackendServiceGrpc(
			services.NewBackend(
				backend.NewFileBackend(f),
				*verbose,
				*maxChunkSize,
			),
		),
	)

	lis, err := net.Listen("tcp", *laddr)
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		<-done

		log.Println("Exiting gracefully")

		_ = lis.Close()
	}()

	log.Println("Listening on", *laddr)

	if err := srv.Serve(lis); err != nil && !utils.IsClosedErr(err) {
		panic(err)
	}
}
