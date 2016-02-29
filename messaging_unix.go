// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"os"

	"golang.org/x/sys/unix"
)

// MessageQueue is a unix-specific ipc mechanism based on message passing
type MessageQueue struct {
	id int
}

type key uint64

func CreateMessageQueue(name string, exclusive bool, perm os.FileMode) (*MessageQueue, error) {
	panic("unimplemented")
}

func mqName(name string) string {
	return os.TempDir() + "/" + name
}

func ftok(name string) key {
	var statfs unix.Stat_t
	if err := unix.Stat(name, &statfs); err != nil {
		return key(0)
	}
	return key(statfs.Ino&0xFFFF | ((statfs.Dev & 0xFF) << 16))
}
