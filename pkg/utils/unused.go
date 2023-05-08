package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoUnusedNBDDeviceFound = errors.New("no unused NBD devices found")
)

func FindUnusedNBDDevice() (string, error) {
	var nbdDevice string
	nbdPaths, err := filepath.Glob("/dev/nbd*")
	if err != nil {
		return "", err
	}

	for _, nbdPath := range nbdPaths {
		f, err := os.OpenFile(nbdPath, os.O_EXCL, 0)
		if err == nil {
			nbdDevice = nbdPath

			if err := f.Close(); err != nil {
				return "", err
			}

			break
		} else if !strings.HasSuffix(err.Error(), "device or resource busy") {
			return "", err
		}
	}

	if nbdDevice == "" {
		return "", ErrNoUnusedNBDDeviceFound
	}

	return nbdDevice, nil
}
