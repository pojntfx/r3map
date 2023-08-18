package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
)

func main() {
	size := flag.Int64("size", 536870912, "Size of the resource in byte")

	flag.Parse()

	f, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(f.Name())

	if err := f.Truncate(*size); err != nil {
		panic(err)
	}

	b := backend.NewFileBackend(f)

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
		b,
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
