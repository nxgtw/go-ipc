// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux,amd64 darwin freebsd

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/nxgtw/go-ipc/internal/allocator"
	"github.com/nxgtw/go-ipc/internal/common"

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
		if err == unix.EEXIST || err == unix.ENOENT {
			return 0, &os.PathError{Op: "SEMGET", Path: "", Err: err}
		}
		return 0, os.NewSyscallError("SEMGET", err)
	}
	return int(id), nil
}

func semctl(id, num, cmd int) error {
	_, _, err := unix.Syscall(sysSemCtl, uintptr(id), uintptr(num), uintptr(cmd))
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
	_, _, err := unix.Syscall(sysSemOp, uintptr(id), uintptr(pOps), uintptr(len(ops)))
	allocator.Use(pOps)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMOP", err)
	}
	return nil
}
