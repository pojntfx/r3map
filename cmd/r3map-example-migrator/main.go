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
	"strings"
	"sync"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	v1 "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	size := flag.Int64("size", 4096*8192*10, "Size of the resource")

	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	maxChunkSize := flag.Int64("max-chunk-size", services.MaxChunkSize, "Maximum chunk size to support")

	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")

	raddr := flag.String("raddr", "", "Remote address (set to enable leeching)")
	laddr := flag.String("laddr", "", "Listen address (set to enable seeding)")

	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(file.Name())

	if err := file.Truncate(*size); err != nil {
		panic(err)
	}

	bar := progressbar.NewOptions(
		int(*size),
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

	migrator := migration.NewPathMigrator(
		ctx,

		backend.NewFileBackend(file),

		&migration.MigratorOptions{
			ChunkSize:    *chunkSize,
			MaxChunkSize: *maxChunkSize,

			PullWorkers: *pullWorkers,

			Verbose: *verbose,
		},
		&migration.MigratorHooks{
			OnBeforeSync: func() error {
				log.Println("Suspending app")

				return nil
			},
			OnAfterSync: func(dirtyOffsets []int64) error {
				bar.Clear()

				delta := (len(dirtyOffsets) * int(*chunkSize))

				log.Printf("Invalidated: %.2f MB (%.2f Mb)", float64(delta)/(1024*1024), (float64(delta)/(1024*1024))*8)

				bar.ChangeMax(int(*size) + delta)

				bar.Describe("Finalizing")

				return nil
			},

			OnBeforeClose: func() error {
				log.Println("Stopping app")

				return nil
			},

			OnChunkIsLocal: func(off int64) error {
				bar.Add(int(*chunkSize))

				return nil
			},
		},

		nil,
		nil,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := migrator.Wait(); err != nil {
			panic(err)
		}
	}()

	var (
		path string
		svc  *services.Seeder
	)
	if strings.TrimSpace(*raddr) != "" {
		conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		log.Println("Connected to", *raddr)

		defer migrator.Close()
		finalize, _, err := migrator.Leech(services.NewLeecherGrpc(v1.NewSeederClient(conn)))
		if err != nil {
			panic(err)
		}

		log.Println("Press <ENTER> to finalize")

		bufio.NewScanner(os.Stdin).Scan()

		seed, p, err := finalize()
		if err != nil {
			panic(err)
		}
		path = p

		bar.Clear()

		log.Println("Resuming app on", path)

		if strings.TrimSpace(*laddr) == "" {
			return
		}

		svc, err = seed()
		if err != nil {
			panic(err)
		}
	}

	if strings.TrimSpace(*laddr) != "" {
		if svc == nil {
			defer migrator.Close()
			path, _, svc, err = migrator.Seed()
			if err != nil {
				panic(err)
			}

			log.Println("Starting app on", path)
		}

		server := grpc.NewServer()

		v1.RegisterSeederServer(server, services.NewSeederGrpc(svc))

		lis, err := net.Listen("tcp", *laddr)
		if err != nil {
			panic(err)
		}
		defer lis.Close()

		log.Println("Listening on", *laddr)

		go func() {
			log.Println("Press <ENTER> to invalidate")

			bufio.NewScanner(os.Stdin).Scan()

			file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			if _, err := io.CopyN(
				file,
				rand.Reader,
				int64(math.Floor(
					float64(*size)*(float64(*invalidate)/float64(100)),
				)),
			); err != nil {
				panic(err)
			}
		}()

		go func() {
			if err := server.Serve(lis); err != nil {
				if !utils.IsClosedErr(err) {
					panic(err)
				}

				return
			}
		}()
	}

	wg.Wait()
}
