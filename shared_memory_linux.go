// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux

package ipc

import (
	"errors"
	"os"
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

func DestroyMemoryObject(name string) error {
	if path, err := shmName(name); err != nil {
		return err
	} else {
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
}

// glibc/sysdeps/posix/shm_open.c
func shmOpen(path string, mode int, perm os.FileMode) (file *os.File, err error) {
	var osMode int
	osMode, err = shmModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	switch {
	case mode&(O_OPEN_ONLY|O_CREATE_ONLY) != 0:
		file, err = os.OpenFile(path, osMode, perm)
	case mode&O_OPEN_OR_CREATE != 0:
		amode, _ := accessModeToOsMode(mode)
		for {
			if file, err = os.OpenFile(path, amode|unix.O_CREAT|unix.O_EXCL, perm); !os.IsExist(err) {
				break
			} else {
				if file, err = os.OpenFile(path, amode, perm); !os.IsNotExist(err) {
					break
				}
			}
		}
	}
	return
}

// glibc/sysdeps/posix/shm-directory.h
func shmName(name string) (string, error) {
	name = strings.TrimLeft(name, "/")
	nameLen := len(name)
	if nameLen == 0 || nameLen >= maxNameLen || strings.Contains(name, "/") {
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
