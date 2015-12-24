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

func shmName(name string) (string, error) {
	if runtime.GOOS == "darwin" {
		name = fmt.Sprintf("%s\t%d", name, syscall.Geteuid())
	}
	return name, nil
}

func shmOpen(path string, mode int, perm os.FileMode) (file *os.File, err error) {
	var osMode int
	osMode, err = shmModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	nameBytes, err := syscall.BytePtrFromString(path)
	if err != nil {
		return nil, err
	}
	bytes := unsafe.Pointer(nameBytes)
	fd, _, err := syscall.Syscall(syscall.SYS_SHM_OPEN, uintptr(bytes), uintptr(osMode), uintptr(perm))
	use(bytes)
	if err != syscall.Errno(0) {
		return nil, err
	}
	return os.NewFile(fd, path), nil
}
