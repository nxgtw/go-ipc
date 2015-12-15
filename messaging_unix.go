// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	ipc_PRIVATE  = 0
	ipc_CREAT    = 01000
	ipc_EXCL     = 02000
	ipc_NOWAIT   = 04000
	ipc_RMID     = 0
	ipc_SET      = 1
	ipc_STAT     = 2
	MSGQUEUE_NEW = ipc_PRIVATE
)

type MessageQueue struct {
	id int
}

func NewMessageQueue(id int, flags int, perm os.FileMode) (*MessageQueue, error) {
	sysflags, err := validateNewMQArgumants(id, flags, perm)
	if err != nil {
		return nil, err
	}
	r1, _, err := syscall.Syscall(unix.SYS_MSGGET, uintptr(id), uintptr(sysflags), 0)
	if err != syscall.Errno(0) {
		return nil, err
	}
	return &MessageQueue{id: int(r1)}, nil
}

func (mq *MessageQueue) Send(typ int, value interface{}) error {
	return nil
}

func (mq *MessageQueue) Id() int {
	return mq.id
}

func (mq *MessageQueue) Destroy() error {
	return DestroyMessageQueue(mq.Id())
}

func validateNewMQArgumants(id int, flags int, perm os.FileMode) (int, error) {
	var sysflags int
	if id == MSGQUEUE_NEW {
		if flags == O_OPEN_ONLY {
			return 0, fmt.Errorf("MSGQUEUE_NEW and O_OPEN_ONLY cannot be used together")
		}
	}
	if flags == O_OPEN_OR_CREATE {
		sysflags |= ipc_CREAT
	} else if flags == O_CREATE_ONLY {
		sysflags |= ipc_EXCL
	} else if flags != O_OPEN_ONLY {
		return 0, fmt.Errorf("invalid open flags")
	}
	return sysflags | int(perm), nil
}

func DestroyMessageQueue(id int) error {
	_, _, err := syscall.Syscall(unix.SYS_MSGCTL, uintptr(id), uintptr(ipc_RMID), 0)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
