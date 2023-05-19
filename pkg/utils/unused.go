package utils

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ErrNoUnusedNBDDeviceFound = errors.New("no unused NBD devices found")
)

func FindUnusedNBDDevice() (string, error) {
	statPaths, err := filepath.Glob(path.Join("/sys", "block", "nbd*", "size"))
	if err != nil {
		return "", err
	}

	for _, statPath := range statPaths {
		rsize, err := os.ReadFile(statPath)
		if err != nil {
			return "", err
		}

		size, err := strconv.ParseInt(strings.TrimSpace(string(rsize)), 10, 64)
		if err != nil {
			return "", err
		}

		if size == 0 {
			return filepath.Join("/dev", filepath.Base(filepath.Dir(statPath))), nil
		}
	}

	return "", ErrNoUnusedNBDDeviceFound
}
