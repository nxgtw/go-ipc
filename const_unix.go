// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

const (
	O_FIFO_NONBLOCK = 0x00000040 // for FIFO open only
)
