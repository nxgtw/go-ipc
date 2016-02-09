// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd netbsd openbsd

package ipc

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
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
	var f *os.File
	opener := func(mode int) error {
		fd, err := shm_open(path, mode, int(perm))
		if err != nil {
			return err
		}
		f = os.NewFile(fd, path)
		return nil
	}
	_, err := openOrCreateFile(opener, mode)
	return f, err
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
