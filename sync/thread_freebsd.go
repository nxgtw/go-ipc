// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

func gettid() (int, error) {
	var tid int64
	tidPtr := unsafe.Pointer(&tid)
	_, _, err := unix.Syscall(unix.SYS_THR_SELF, uintptr(tidPtr), uintptr(0), uintptr(0))
	allocator.Use(tidPtr)
	if err != syscall.Errno(0) {
		return 0, err
	}
	return int(tid), nil
}

func killThread(tid int) error {
	_, _, err := unix.Syscall(unix.SYS_THR_KILL, uintptr(tid), uintptr(unix.SIGUSR2), 0)
	return err
}
