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
	fd   int
	path string
}

func newMemoryRegionImpl(name string, size int64, mode int, flags uint32) (*memoryRegionImpl, error) {
	var err error
	mode, err = modeToUnixMode(mode)
	if err != nil {
		return nil, err
	}
	flags, err = flagsToUnixFlags(flags)
	if err != nil {
		return nil, err
	}
	var path string
	if path, err = shmName(name); err != nil {
		return nil, err
	}
	fd, err := shmOpen(path, mode, flags)
	if err != nil {
		return nil, err
	}
	if err = unix.Ftruncate(fd, size); err != nil {
		return nil, err
	}
	return &memoryRegionImpl{fd: fd, path: path}, nil
}

func (impl *memoryRegionImpl) Destroy() {
	impl.Close()
	unix.Unlink(impl.path) // TODO (avd) - error handling
}

func (impl *memoryRegionImpl) Close() {
	unix.Close(impl.fd) // TODO (avd) - error handling
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, flag int, mode uint32) (fd int, err error) {
	if fd, err = unix.Open(path, flag, 0777); err != nil { // TODO (avd) - 0777?
		return
	}
	unix.CloseOnExec(fd)
	return
}

// glibc/sysdeps/posix/shm-directory.h
func shmName(name string) (string, error) {
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

func flagsToUnixFlags(flags uint32) (uint32, error) {
	var uflags uint32
	if flags&SHM_CREATE != 0 {
		uflags |= unix.O_CREAT
	}
	if flags&SHM_CREATE_IF_NOT_EXISTS != 0 {
		uflags |= unix.O_EXCL
	}
	if flags&SHM_TRUNC != 0 {
		uflags |= unix.O_TRUNC
	}
	if flags&SHM_RDONLY != 0 {
		uflags |= unix.O_RDONLY
	}
	if flags&SHM_RDWR != 0 {
		uflags |= unix.O_RDWR
	}
	return flags, nil
}

func modeToUnixMode(mode int) (int, error) {
	return unix.O_RDWR | unix.O_CREAT, nil
}
