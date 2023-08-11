package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	v1 "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	size := flag.Int64("size", 536870912, "Size of the resource in byte")

	raddr := flag.String("raddr", "localhost:1337", "Remote address")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Println("Connected to", *raddr)

	devPath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		panic(err)
	}

	devFile, err := os.Open(devPath)
	if err != nil {
		panic(err)
	}
	defer devFile.Close()

	mnt := mount.NewDirectFileMount(
		lbackend.NewRPCBackend(
			ctx,
			services.NewBackendRemoteGrpc(
				v1.NewBackendClient(conn),
			),
			*size,
			*verbose,
		),
		devFile,

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
