package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
)

func main() {
	laddr := flag.String("laddr", ":1337", "Listen address")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	size := flag.Int64("size", 671088640, "Size of the file to allocate")
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	maxChunkSize := flag.Int64("max-chunk-size", services.MaxChunkSize, "Maximum chunk size to support")

	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate in between Track() and Finalize()")

	flag.Parse()

	file, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(file.Name())

	if err := file.Truncate(*size); err != nil {
		panic(err)
	}

	b := backend.NewFileBackend(file)

	seeder := migration.NewFileSeeder(
		b,

		&migration.SeederOptions{
			ChunkSize:    *chunkSize,
			MaxChunkSize: *maxChunkSize,

			Verbose: *verbose,
		},
		&migration.SeederHooks{
			OnBeforeSync: func() error {
				log.Println("Suspending app")

				return nil
			},
			OnBeforeClose: func() error {
				log.Println("Stopping app")

				return nil
			},
		},

		nil,
		nil,
	)

	errs := make(chan error)
	go func() {
		if err := seeder.Wait(); err != nil {
			errs <- err

			return
		}

		close(errs)
	}()

	defer seeder.Close()
	deviceFile, svc, err := seeder.Open()
	if err != nil {
		panic(err)
	}

	log.Println("Connected on", deviceFile.Name())

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

	go func() {
		log.Println("Press <ENTER> to invalidate")

		bufio.NewScanner(os.Stdin).Scan()

		beforeInvalidate := time.Now()
		if _, err := io.CopyN(
			deviceFile,
			rand.Reader,
			int64(math.Floor(
				float64(*size)*(float64(*invalidate)/float64(100)),
			)),
		); err != nil {
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
