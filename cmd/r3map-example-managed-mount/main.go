package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	v1 "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	size := flag.Int64("size", 4096*8192*10, "Size of the resource")

	raddr := flag.String("raddr", "localhost:1337", "Remote address")

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

	conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Println("Connected to", *raddr)

	mnt := mount.NewManagedFileMount(
		ctx,

		lbackend.NewRPCBackend(
			ctx,
			services.NewBackendRemoteGrpc(
				v1.NewBackendClient(conn),
			),
			false,
		),
		backend.NewFileBackend(f),

		&mount.ManagedMountOptions{
			Verbose: *verbose,
		},
		nil,

		nil,
		nil,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := mnt.Wait(); err != nil {
			panic(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		<-done

		log.Println("Exiting gracefully")

		_ = mnt.Close()
	}()

	defer mnt.Close()
	file, err := mnt.Open()
	if err != nil {
		panic(err)
	}

	log.Println("Resource available on", file.Name())

	wg.Wait()
}
