package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/r3map/pkg/migration"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

type rpcReaderAt struct {
	ctx context.Context

	remote *services.SourceRemote

	verbose bool
}

func (b *rpcReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if b.verbose {
		log.Printf("ReadAt(len(p) = %v, off = %v)", len(p), off)
	}

	r, err := b.remote.ReadAt(b.ctx, len(p), off)
	if err != nil {
		return -1, err
	}

	n = r.N
	copy(p, r.P)

	return
}

var (
	errNoPeerFound = errors.New("no peer found")
)

func main() {
	raddr := flag.String("raddr", "localhost:1337", "Remote address")

	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	pullWorkers := flag.Int64("pull-workers", 512, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	registry := rpc.NewRegistry(
		&struct{}{},
		services.SourceRemote{},

		time.Second*10,
		ctx,
		&rpc.Options{
			ResponseBufferLen: rpc.DefaultResponseBufferLen,
			OnClientConnect: func(remoteID string) {
				ready <- struct{}{}
			},
		},
	)

	conn, err := net.Dial("tcp", *raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		if err := registry.Link(conn); err != nil {
			if !utils.IsClosedErr(err) {
				panic(err)
			}
		}
	}()

	<-ready

	log.Println("Connected to", conn.RemoteAddr())

	var peer *services.SourceRemote
	for _, candidate := range registry.Peers() {
		peer = &candidate

		break
	}

	if peer == nil {
		panic(errNoPeerFound)
	}

	size, err := peer.Size(ctx)
	if err != nil {
		panic(err)
	}

	local := backend.NewMemoryBackend(make([]byte, size))

	output := backend.NewMemoryBackend(make([]byte, size))

	mnt := migration.NewDestination(
		ctx,

		&rpcReaderAt{
			ctx,
			peer,
			*verbose,
		},
		size,

		func() ([]int64, error) {
			return peer.Flush(ctx)
		},

		local,

		&migration.Options{
			ChunkSize: *chunkSize,

			PullWorkers: *pullWorkers,

			Verbose: *verbose,
		},

		nil,
		nil,
	)

	go func() {
		if err := mnt.Wait(); err != nil {
			panic(err)
		}
	}()

	devicePath, err := mnt.Open()
	if err != nil {
		panic(err)
	}
	defer mnt.Close()

	deviceFile, err := os.OpenFile(devicePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer deviceFile.Close()

	go func() {
		log.Println("Press <ENTER> to finalize pull")

		bufio.NewScanner(os.Stdin).Scan()

		log.Println("Finalizing pull")

		if err := mnt.FinalizePull(); err != nil {
			panic(err)
		}
	}()

	if _, err := io.CopyN(
		io.NewOffsetWriter(
			output,
			0,
		), deviceFile, size); err != nil {
		panic(err)
	}
}
