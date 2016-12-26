// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

func semtimedop(id int, ops []sembuf, timeout *unix.Timespec) error {
	if len(ops) == 0 {
		return nil
	}
	pOps := unsafe.Pointer(&ops[0])
	pTimeout := unsafe.Pointer(timeout)
	_, _, err := unix.Syscall6(unix.SYS_SEMTIMEDOP, uintptr(id), uintptr(pOps), uintptr(len(ops)), uintptr(pTimeout), 0, 0)
	allocator.Use(pOps)
	allocator.Use(pTimeout)
	if err != syscall.Errno(0) {
		return os.NewSyscallError("SEMTIMEDOP", err)
	}
	return nil
}
