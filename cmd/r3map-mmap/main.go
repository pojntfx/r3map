package main

import (
	"flag"
	"io"
	"log"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	file := flag.String("file", "/dev/nbd0", "Path to device file to mmap")

	flag.Parse()

	f, err := os.OpenFile(*file, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		panic(err)
	}

	b, err := syscall.Mmap(
		int(f.Fd()),
		0,
		int(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		panic(err)
	}
	defer syscall.Munmap(b)

	defer func() {
		_, _, _ = syscall.Syscall(
			syscall.SYS_MSYNC,
			uintptr(unsafe.Pointer(&b[0])),
			uintptr(len(b)),
			uintptr(syscall.MS_SYNC),
		)
	}()

	log.Println("Connected to", *file)

	log.Printf("First byte: %v", string(b[0]))

	log.Println("Writing a to first byte")

	b[0] = 'a'

	log.Printf("First byte: %v", string(b[0]))
}
