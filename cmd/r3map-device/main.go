package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/device"
)

func main() {
	file := flag.String("file", "/dev/nbd0", "Path to device file to use")
	size := flag.Int64("size", 1073741824, "Size of the memory region to expose")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	b := backend.NewMemoryBackend(make([]byte, *size), *verbose)
	f, err := os.Open(*file)
	if err != nil {
		panic(err)
	}

	d := device.NewDevice(
		b,
		f,

		nil,
		nil,
	)

	if err := d.Open(); err != nil {
		panic(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		<-sigCh

		if *verbose {
			log.Println("Cleaning up resources")
		}

		go func() {
			<-sigCh

			log.Fatal("Force exiting")
		}()

		if err := d.Close(); err != nil {
			panic(err)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}

		os.Exit(0)
	}()

	log.Println("Listening on", *file)

	if err := d.Wait(); err != nil {
		panic(err)
	}
}
