// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,amd64 darwin, freebsd

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

var (
	sysSemGet uintptr
	sysSemCtl uintptr
	sysSemOp  uintptr
)

func semget(k common.Key, nsems, semflg int) (int, error) {
	id, _, err := unix.Syscall(sysSemGet, uintptr(k), uintptr(nsems), uintptr(semflg))
	if err != syscall.Errno(0) {
		return 0, err
	}
	return int(id), nil
}

func semctl(id, num, cmd int) error {
	_, _, err := syscall.Syscall(sysSemCtl, uintptr(id), uintptr(num), uintptr(cmd))
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMCTL", err)
	}
	return nil
}

func semop(id int, ops []sembuf) error {
	if len(ops) == 0 {
		return nil
	}
	pOps := unsafe.Pointer(&ops[0])
	_, _, err := syscall.Syscall(sysSemOp, uintptr(id), uintptr(pOps), uintptr(len(ops)))
	allocator.Use(pOps)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMOP", err)
	}
	return nil
}
