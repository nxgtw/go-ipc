// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import "golang.org/x/sys/unix"

func mach_thread_self() uint32

func gettid() (uint32, error) {
	return mach_thread_self(), nil
}

func killThread(port uint32) error {
	_, _, err := unix.Syscall(unix.SYS___PTHREAD_KILL, uintptr(port), uintptr(unix.SIGUSR2), 0)
	return err
}
