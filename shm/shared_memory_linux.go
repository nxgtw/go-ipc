// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux

package shm

import (
	"errors"
	"os"
	"strings"
	"sync"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	maxNameLen     = 255
	defaultShmPath = "/dev/shm/"
)

var (
	shmPathOnce sync.Once
	shmPath     string
)

func destroyMemoryObject(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, mode int, perm os.FileMode) (*os.File, error) {
	osMode, err := common.OpenModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, osMode, perm)
	return file, err
}

// glibc/sysdeps/posix/shm-directory.h
func shmName(name string) (string, error) {
	name = strings.TrimLeft(name, "/")
	nameLen := len(name)
	if nameLen == 0 || nameLen >= maxNameLen || strings.Contains(name, "/") {
		return "", errors.New("invalid shm name")
	}
	var dir string
	var err error
	if dir, err = shmDirectory(); err != nil {
		return "", err
	}
	return dir + name, nil
}

func shmDirectory() (string, error) {
	shmPathOnce.Do(locateShmFs)
	if len(shmPath) == 0 {
		return shmPath, errors.New("error locating the shared memory path")
	}
	return shmPath, nil
}
