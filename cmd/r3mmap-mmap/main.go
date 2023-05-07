package main

import (
	"flag"
	"log"
	"os"
	"syscall"
)

func main() {
	dev := flag.String("dev", "/dev/nbd0", "Device to mmap")

	flag.Parse()

	f, err := os.OpenFile(*dev, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		panic(err)
	}

	b, err := syscall.Mmap(int(f.Fd()), 0, int(s.Size()), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}

	log.Println(b)
}
