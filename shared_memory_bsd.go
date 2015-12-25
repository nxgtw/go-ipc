// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd

package ipc

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func destroyMemoryObject(path string) error {
	err := shm_unlink(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

func shmName(name string) (string, error) {
	// workaround from http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if runtime.GOOS == "darwin" {
		name = fmt.Sprintf("%s\t%d", name, syscall.Geteuid())
	}
	return "/tmp/" + name, nil
}

func shmOpen(path string, mode int, perm os.FileMode) (*os.File, error) {
	var fd uintptr
	var err error
	switch {
	case mode&(O_OPEN_ONLY|O_CREATE_ONLY) != 0:
		var osMode int
		osMode, err = shmModeToOsMode(mode)
		if err != nil {
			return nil, err
		}
		fd, err = shm_open(path, osMode, int(perm))
	case mode&O_OPEN_OR_CREATE != 0:
		amode, _ := accessModeToOsMode(mode)
		for {
			if fd, err = shm_open(path, amode|unix.O_CREAT|unix.O_EXCL, int(perm)); !os.IsExist(err) {
				break
			} else {
				if fd, err = shm_open(path, amode, int(perm)); !os.IsNotExist(err) {
					break
				}
			}
		}
	default:
		err = fmt.Errorf("unknown open mode")
	}
	if err != nil {
		return nil, err
	} else {
		return os.NewFile(fd, path), nil
	}
}

// syscalls

func shm_open(name string, flags, mode int) (uintptr, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	bytes := unsafe.Pointer(nameBytes)
	fd, _, err := syscall.Syscall(syscall.SYS_SHM_OPEN, uintptr(bytes), uintptr(flags), uintptr(mode))
	use(bytes)
	if err != syscall.Errno(0) {
		return 0, err
	}
	return fd, nil
}

func shm_unlink(name string) error {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	bytes := unsafe.Pointer(nameBytes)
	_, _, err = syscall.Syscall(syscall.SYS_SHM_UNLINK, uintptr(bytes), uintptr(0), uintptr(0))
	use(bytes)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
