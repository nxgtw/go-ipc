// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux

package ipc

import (
	"errors"
	"os"
	"strings"
	"sync"
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
	var file *os.File
	opener := func(mode int) error {
		var err error
		file, err = os.OpenFile(path, mode, perm)
		return err
	}
	_, err := openOrCreateFile(opener, mode)
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
