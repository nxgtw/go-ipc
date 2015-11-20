// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

const (
	maxNameLen     = 255
	defaultShmPath = "/dev/shm/"
)

var (
	shmPathOnce sync.Once
	shmPath     string
)

type memoryRegionImpl struct {
	fd int
}

func newMemoryRegionImpl() *memoryRegionImpl {
	return &memoryRegionImpl{}
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(name string, flag int, mode uint32) (err error) {
	var path string
	if path, err = shmName(); err != nil {
		return err
	}
	var fd int
	if fd, err = unix.Open(path, flag, mode); err != nil {
		return err
	}
	unix.CloseOnExec(fd)
	return nil
}

// glibc/sysdeps/posix/shm-directory.h
func shmName() (name string, err error) {
	name = strings.TrimLeft(name, "/")
	nameLen := len(name)
	if nameLen == 1 || nameLen >= maxNameLen || strings.Contains(name, "/") {
		return "", errors.New("invalid shm name")
	}
	if dir, err := shmDirectory(); err != nil {
		return "", err
	} else {
		return dir + name, nil
	}
}

func shmDirectory() (string, error) {
	shmPathOnce.Do(locateShmFs)
	if len(shmPath) == 0 {
		return shmPath, errors.New("error locating the shared memory path")
	}
	return shmPath, nil
}
