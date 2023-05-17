package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	"github.com/pojntfx/dudirekta/pkg/rpc"
	"github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
)

type sliceRwat struct {
	data []byte
	rtt  time.Duration
}

func (rw *sliceRwat) ReadAt(p []byte, off int64) (n int, err error) {
	if rw.rtt > 0 {
		time.Sleep(rw.rtt)
	}

	if off >= int64(len(rw.data)) {
		return 0, io.EOF
	}

	n = copy(p, rw.data[off:off+int64(len(p))])

	return n, nil
}

func (rw *sliceRwat) WriteAt(p []byte, off int64) (n int, err error) {
	if rw.rtt > 0 {
		time.Sleep(rw.rtt)
	}

	if off >= int64(len(rw.data)) {
		return 0, io.ErrShortWrite
	}

	n = copy(rw.data[off:off+int64(len(p))], p)
	if n < len(p) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

func allocateSlice(size int) ([]byte, func() error, error) {
	p, err := syscall.Mmap(
		-1,
		0,
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
	if err != nil {
		return []byte{}, nil, nil
	}

	return p, func() error {
		_, _, _ = syscall.Syscall(
			syscall.SYS_MSYNC,
			uintptr(unsafe.Pointer(&p[0])),
			uintptr(len(p)),
			uintptr(syscall.MS_SYNC),
		)

		if err := syscall.Munmap(p); err != nil {
			panic(err)
		}

		return nil
	}, nil
}

func main() {
	chunkSize := flag.Int64("chunk-size", 4096, "Chunk size to use")
	pullWorkers := flag.Int64("pull-workers", 512, "Puller workers to launch in the background; pass in 0 to disable preemptive pull")
	pushWorkers := flag.Int64("push-workers", 512, "Push workers to launch in the background; pass in 0 to disable push")
	completePull := flag.Bool("complete-pull", false, "Whether to completely pull the remote to the local slice before starting benchmark")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")
	check := flag.Bool("check", true, "Check if local and remote hashes match")
	localRTT := flag.Duration("local-rtt", 0, "Simulated RTT of the local slice")
	pusherInterval := flag.Duration("pusher-interval", 5*time.Minute, "Interval after which to push chunks to the remote")
	raddr := flag.String("raddr", "localhost:1337", "Remote address")

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
						// Create remote and local backends
						rpcBackend := backend.NewRPCBackend(ctx, peer, *verbose)

						size, err := rpcBackend.Size()
						if err != nil {
							return err
						}
						chunkCount := size / *chunkSize

						// Create local slice
						localSlice, freeLocalSlice, err := allocateSlice(int(size))
						if err != nil {
							panic(err)
						}
						defer freeLocalSlice()

						localFile := &sliceRwat{localSlice, *localRTT}

						// Setup the device
						path, err := utils.FindUnusedNBDDevice(time.Millisecond * 50)
						if err != nil {
							panic(err)
						}

						df, err := os.Open(path)
						if err != nil {
							panic(err)
						}
						defer df.Close()

						remote := chunks.NewChunkedReadWriterAt(rpcBackend, *chunkSize, chunkCount)

						var local chunks.ReadWriterAt
						if *pushWorkers > 0 {
							l := chunks.NewChunkedReadWriterAt(localFile, *chunkSize, chunkCount)

							// Setup the pusher
							pusher := chunks.NewPusher(
								ctx,
								l,
								remote,
								*chunkSize,
								*pusherInterval,
							)

							go func() {
								if err := pusher.Wait(); err != nil {
									log.Println("Fatal error while waiting on pusher:", err)

									if *verbose {
										debug.PrintStack()
									}

									os.Exit(1)
								}
							}()

							if err := pusher.Open(*pushWorkers); err != nil {
								panic(err)
							}
							defer pusher.Close()

							sigCh := make(chan os.Signal, 1)
							signal.Notify(sigCh, syscall.SIGHUP)
							defer close(sigCh)

							go func() {
								for range sigCh {
									before := time.Now()

									n, err := pusher.Flush()
									if err != nil {
										log.Println("Could not flush:", err)

										continue
									}

									if err := rpcBackend.Sync(); err != nil {
										log.Println("Could not sync:", err)

										continue
									}

									after := time.Since(before)

									log.Printf("Manually flushed and synced %v chunks in %v", n, after)
								}
							}()

							local = pusher
						} else {
							local = chunks.NewChunkedReadWriterAt(localFile, *chunkSize, chunkCount)
						}

						srw := chunks.NewSyncedReadWriterAt(remote, local, func(off int64) error {
							if *pushWorkers > 0 {
								if err := local.(*chunks.Pusher).MarkOffsetPushable(off); err != nil {
									return err
								}
							}

							return nil
						})

						if *pullWorkers > 0 {
							// Setup the puller
							puller := chunks.NewPuller(
								ctx,
								srw,
								*chunkSize,
								chunkCount,
								func(offset int64) int64 {
									return 1
								},
							)

							if !*completePull {
								go func() {
									if err := puller.Wait(); err != nil {
										log.Println("Fatal error while waiting on puller:", err)

										if *verbose {
											debug.PrintStack()
										}

										os.Exit(1)
									}
								}()
							}

							if err := puller.Open(*pullWorkers); err != nil {
								panic(err)
							}
							defer puller.Close()

							if *completePull {
								if err := puller.Wait(); err != nil {
									panic(err)
								}
							}
						}

						arw := chunks.NewArbitraryReadWriterAt(srw, *chunkSize)

						remoteBackend := backend.NewReaderAtBackend(
							arw,
							func() (int64, error) {
								return size, nil
							},
							func() error {
								if *pushWorkers > 0 {
									beforeFlush := time.Now()

									n, err := local.(*chunks.Pusher).Flush()
									if err != nil {
										return err
									}

									if err := rpcBackend.Sync(); err != nil {
										return err
									}

									afterFlush := time.Since(beforeFlush)

									fmt.Printf("Flush and sync: %v chunks in %v\n", n, afterFlush)
								}

								return nil
							},
							*verbose,
						)

						d := device.NewDevice(
							remoteBackend,
							df,

							nil,
							nil,
						)

						go func() {
							if err := d.Wait(); err != nil {
								log.Println("Fatal error while waiting on device:", err)

								if *verbose {
									debug.PrintStack()
								}

								os.Exit(1)
							}
						}()

						if err := d.Open(); err != nil {
							panic(err)
						}
						defer d.Close()

						cf, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
						if err != nil {
							panic(err)
						}
						defer cf.Close()

						// Setup the r3mapped and output slices
						r3mappedSlice, err := syscall.Mmap(
							int(cf.Fd()),
							0,
							int(size),
							syscall.PROT_READ|syscall.PROT_WRITE,
							syscall.MAP_SHARED,
						)
						if err != nil {
							panic(err)
						}
						defer syscall.Munmap(r3mappedSlice)

						defer func() {
							_, _, _ = syscall.Syscall(
								syscall.SYS_MSYNC,
								uintptr(unsafe.Pointer(&r3mappedSlice[0])),
								uintptr(len(r3mappedSlice)),
								uintptr(syscall.MS_SYNC),
							)
						}()

						outputSlice, freeOutputSlice, err := allocateSlice(int(size))
						if err != nil {
							panic(err)
						}
						defer freeOutputSlice()

						// Run the benchmark
						beforeRead := time.Now()

						copy(outputSlice, r3mappedSlice)

						afterRead := time.Since(beforeRead)

						fmt.Printf("Read: %.2f MB/s\n", float64(size)/(1024*1024)/afterRead.Seconds())

						// Validate the results
						validateResults := func() error {
							remoteHash := xxhash.New()
							if _, err := io.Copy(remoteHash, io.NewSectionReader(rpcBackend, 0, size)); err != nil {
								return err
							}

							localHash := xxhash.New()
							if _, err := io.Copy(localHash, bytes.NewReader(localSlice)); err != nil {
								return err
							}

							r3mappedHash := xxhash.New()
							if _, err := io.Copy(r3mappedHash, bytes.NewReader(r3mappedSlice)); err != nil {
								return err
							}

							outputHash := xxhash.New()
							if _, err := io.Copy(outputHash, bytes.NewReader(outputSlice)); err != nil {
								return err
							}

							if remoteHash.Sum64() != localHash.Sum64() {
								return errors.New("remote, local, r3mapped and output hashes don't match")
							}

							fmt.Println("Check: Remote, local, r3mapped and output hashes match.")

							return nil
						}

						if *check {
							if err := validateResults(); err != nil {
								panic(err)
							}
						}

						beforeWrite := time.Now()

						if _, err := rand.Read(r3mappedSlice); err != nil {
							panic(err)
						}

						_, _, _ = syscall.Syscall(
							syscall.SYS_MSYNC,
							uintptr(unsafe.Pointer(&r3mappedSlice[0])),
							uintptr(len(r3mappedSlice)),
							uintptr(syscall.MS_SYNC),
						)

						afterWrite := time.Since(beforeWrite)

						fmt.Printf("Write: %.2f MB/s\n", float64(size)/(1024*1024)/afterWrite.Seconds())

						if *pushWorkers > 0 {
							if err := remoteBackend.Sync(); err != nil {
								panic(err)
							}

							if *check {
								if err := validateResults(); err != nil {
									panic(err)
								}
							}
						}

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
