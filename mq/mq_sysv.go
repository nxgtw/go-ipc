// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package mq

import (
	"errors"
	"os"
	"unsafe"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	cDefaultMessageType = 1
	cSysVAnyMessage     = 0

	typeDataSize = int(unsafe.Sizeof(int(0)))
)

// SystemVMessageQueue is a System V ipc mechanism based on message passing.
type SystemVMessageQueue struct {
	flags int
	id    int
	name  string
}

// msqidDs is for msgctl syscall, but it is not currently used
type msqidDs struct {
}

// this is to ensure, that system V implementation of ipc mq
// satisfies the minimal queue interface
var (
	_ Messenger = (*SystemVMessageQueue)(nil)
)

// CreateSystemVMessageQueue creates new queue with the given name and permissions.
// 'execute' permission cannot be used.
func CreateSystemVMessageQueue(name string, perm os.FileMode) (*SystemVMessageQueue, error) {
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, err
	}
	id, err := msgget(k, int(perm)|common.IpcCreate|common.IpcExcl)
	if err != nil {
		return nil, err
	}
	return &SystemVMessageQueue{id: id, name: name}, nil
}

// OpenSystemVMessageQueue opens existing message queue.
func OpenSystemVMessageQueue(name string, flags int) (*SystemVMessageQueue, error) {
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, err
	}
	id, err := msgget(k, 0)
	if err != nil {
		return nil, err
	}
	result := &SystemVMessageQueue{id: id, name: name}
	if flags&ipc.O_NONBLOCK != 0 {
		result.flags |= common.IpcNoWait
	}
	return result, nil
}

// Send sends a message.
// It blocks if the queue is full.
func (mq *SystemVMessageQueue) Send(data []byte) error {
	f := func() error { return msgsnd(mq.id, cDefaultMessageType, data, mq.flags) }
	err := common.UninterruptedSyscall(f)
	if err != nil && mq.flags&common.IpcNoWait != 0 && common.IsTimeoutErr(err) {
		err = nil
	}
	return err
}

// Receive receives a message.
// It blocks if the queue is empty.
func (mq *SystemVMessageQueue) Receive(data []byte) error {
	f := func() error { return msgrcv(mq.id, data, cSysVAnyMessage, mq.flags) }
	return common.UninterruptedSyscall(f)
}

// Destroy closes the queue and removes it permanently.
func (mq *SystemVMessageQueue) Destroy() error {
	mq.Close()
	err := msgctl(mq.id, common.IpcRmid, nil)
	if err == nil {
		if err = os.Remove(common.TmpFilename(mq.name)); os.IsNotExist(err) {
			err = nil
		}
	} else if os.IsNotExist(err) {
		err = nil
	}
	return err
}

// Close closes the queue.
// As there is no need to close SystemV mq, this function returns nil.
// It was added to satisfy io.Closer
func (mq *SystemVMessageQueue) Close() error {
	return nil
}

// SetBlocking sets whether the send/receive operations on the queue block.
func (mq *SystemVMessageQueue) SetBlocking(block bool) error {
	if block {
		mq.flags &= ^common.IpcNoWait
	} else {
		mq.flags |= common.IpcNoWait
	}
	return nil
}

// DestroySystemVMessageQueue permanently removes queue with a given name.
func DestroySystemVMessageQueue(name string) error {
	mq, err := OpenSystemVMessageQueue(name, 0)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	return mq.Destroy()
}
