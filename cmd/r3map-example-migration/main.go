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
	"os/signal"
	"strings"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/client"
	v1 "github.com/pojntfx/r3map/pkg/api/proto/migration/v1"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	size := flag.Int64("size", 536870912, "Size of the resource")

	raddr := flag.String("raddr", "", "Remote address (set to enable leeching)")
	laddr := flag.String("laddr", "", "Listen address (set to enable seeding)")

	invalidate := flag.Int("invalidate", 0, "Percentage of chunks (0-100) to invalidate")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(f.Name())

	if err := f.Truncate(*size); err != nil {
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

	bar.Add(client.MaximumBlockSize)

	mgr := migration.NewFileMigrator(
		ctx,

		backend.NewFileBackend(f),

		&migration.MigratorOptions{
			Verbose: *verbose,
		},
		&migration.MigratorHooks{
			OnBeforeSync: func() error {
				log.Println("Suspending app")

				return nil
			},
			OnAfterSync: func(dirtyOffsets []int64) error {
				bar.Clear()

				delta := (len(dirtyOffsets) * client.MaximumBlockSize)

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
				bar.Add(client.MaximumBlockSize)

				return nil
			},
		},

		nil,
		nil,
	)

	finished := make(chan struct{})
	go func() {
		defer close(finished)

		if err := mgr.Wait(); err != nil {
			panic(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		<-done

		log.Println("Exiting gracefully")

		_ = mgr.Close()
	}()

	var (
		file *os.File
		svc  *services.SeederService
	)
	if strings.TrimSpace(*raddr) != "" {
		conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		log.Println("Leeching from", *raddr)

		defer mgr.Close()
		finalize, err := mgr.Leech(services.NewSeederRemoteGrpc(v1.NewSeederClient(conn)))
		if err != nil {
			panic(err)
		}

		log.Println("Press <ENTER> to finalize migration")

		continueCh := make(chan struct{})
		go func() {
			bufio.NewScanner(os.Stdin).Scan()

			continueCh <- struct{}{}
		}()

		select {
		case <-continueCh:
		case <-finished:
			return
		}

		seed, f, err := finalize()
		if err != nil {
			panic(err)
		}
		file = f

		bar.Clear()

		log.Println("Resuming app on", file.Name())

		svc, err = seed()
		if err != nil {
			panic(err)
		}
	}

	if strings.TrimSpace(*laddr) != "" {
		if svc == nil {
			defer mgr.Close()
			file, svc, err = mgr.Seed()
			if err != nil {
				panic(err)
			}

			log.Println("Starting app on", file.Name())
		}

		server := grpc.NewServer()

		v1.RegisterSeederServer(server, services.NewSeederServiceGrpc(svc))

		lis, err := net.Listen("tcp", *laddr)
		if err != nil {
			panic(err)
		}
		defer lis.Close()

		log.Println("Seeding on", *laddr)

		go func() {
			log.Println("Press <ENTER> to invalidate resource")

			bufio.NewScanner(os.Stdin).Scan()

			log.Println("Invalidating resource")

			if _, err := file.Seek(0, io.SeekStart); err != nil {
				panic(err)
			}

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

	<-finished
}
