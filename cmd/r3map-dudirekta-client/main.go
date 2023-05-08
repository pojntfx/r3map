package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

func main() {
	raddr := flag.String("raddr", "localhost:1337", "Remote address")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clients := 0

	registry := rpc.NewRegistry(
		&struct{}{},
		services.BackendRemote{},

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

	go func() {
		log.Println(`Enter one of the following letters followed by <ENTER> to run a function on the remote(s):

- a: Start test of remote memory server`)

		stdin := bufio.NewReader(os.Stdin)

		for {
			line, err := stdin.ReadString('\n')
			if err != nil {
				panic(err)
			}

			for peerID, peer := range registry.Peers() {
				log.Println("Calling functions for peer with ID", peerID)

				switch line {
				case "a\n":
					if err := func() error {
						path, err := utils.FindUnusedNBDDevice()
						if err != nil {
							return err
						}

						b := backend.NewRPCBackend(ctx, peer, *verbose)
						df, err := os.Open(path)
						if err != nil {
							return err
						}
						defer df.Close()

						d := device.NewDevice(
							b,
							df,

							nil,
							nil,
						)

						if err := d.Open(); err != nil {
							return err
						}
						defer d.Close()

						go func() {
							if err := d.Wait(); err != nil {
								panic(err)
							}
						}()

						cf, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
						if err != nil {
							return err
						}
						defer cf.Close()

						size, err := cf.Seek(0, io.SeekEnd)
						if err != nil {
							return err
						}

						p, err := syscall.Mmap(
							int(cf.Fd()),
							0,
							int(size),
							syscall.PROT_READ|syscall.PROT_WRITE,
							syscall.MAP_SHARED,
						)
						if err != nil {
							return err
						}
						defer syscall.Munmap(p)

						defer func() {
							_, _, _ = syscall.Syscall(
								syscall.SYS_MSYNC,
								uintptr(unsafe.Pointer(&p[0])),
								uintptr(len(p)),
								uintptr(syscall.MS_SYNC),
							)
						}()

						log.Printf("First byte: %v", string(p[0]))

						log.Println("Writing a to first byte")

						p[0] = 'a'

						log.Printf("First byte: %v", string(p[0]))

						return nil
					}(); err != nil {
						panic(err)
					}
				default:
					log.Printf("Unknown letter %v, ignoring input", line)

					continue
				}
			}
		}
	}()

	conn, err := net.Dial("tcp", *raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Println("Connected to", conn.RemoteAddr())

	if err := registry.Link(conn); err != nil {
		panic(err)
	}
}
