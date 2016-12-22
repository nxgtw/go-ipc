// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func mach_reply_port() uint32

func killThread(port uint32, signal syscall.Signal) error {
	_, _, err := unix.Syscall(unix.SYS___PTHREAD_KILL, uintptr(port), uintptr(signal), 0)
	return err
}
