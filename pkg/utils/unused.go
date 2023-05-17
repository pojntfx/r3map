package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNoUnusedNBDDeviceFound = errors.New("no unused NBD devices found")
)

func FindUnusedNBDDevice(timeout time.Duration) (string, error) {
	var nbdDevice string
	nbdPaths, err := filepath.Glob("/dev/nbd*")
	if err != nil {
		return "", err
	}

	for _, nbdPath := range nbdPaths {
		errCh := make(chan error, 1)
		go func() {
			f, err := os.OpenFile(nbdPath, os.O_EXCL, 0)
			if err != nil {
				errCh <- err

				return
			}

			errCh <- f.Close()
		}()

		select {
		case err = <-errCh:
			if err == nil {
				nbdDevice = nbdPath
			} else if !strings.HasSuffix(err.Error(), "device or resource busy") {
				return "", err
			}
		case <-time.After(timeout):
			// Consider device to be busy
		}

		if nbdDevice != "" {
			break
		}
	}

	if nbdDevice == "" {
		return "", ErrNoUnusedNBDDeviceFound
	}

	return nbdDevice, nil
}
