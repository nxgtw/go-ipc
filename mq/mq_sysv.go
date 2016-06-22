// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package mq

import (
	"os"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"

	"github.com/pkg/errors"
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
//	name - unique mq name.
//	flag - flag is a combination of os.O_EXCL and O_NONBLOCK.
//	perm - object's permission bits.
func CreateSystemVMessageQueue(name string, flag int, perm os.FileMode) (*SystemVMessageQueue, error) {
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a key")
	}
	sysFlags := int(perm) | common.IpcCreate
	if flag&os.O_EXCL != 0 {
		sysFlags |= common.IpcExcl
	}
	id, err := msgget(k, sysFlags)
	if err != nil {
		return nil, errors.Wrap(err, "msgget failed")
	}
	return &SystemVMessageQueue{id: id, name: name, flags: flag}, nil
}

// OpenSystemVMessageQueue opens existing message queue.
//	name - unique mq name.
//	flag - 0 and O_NONBLOCK.
func OpenSystemVMessageQueue(name string, flags int) (*SystemVMessageQueue, error) {
	k, err := common.KeyForName(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a key")
	}
	id, err := msgget(k, 0)
	if err != nil {
		return nil, errors.Wrap(err, "msgget failed")
	}
	result := &SystemVMessageQueue{id: id, name: name, flags: flags}
	return result, nil
}

// Send sends a message. It blocks if the queue is full.
func (mq *SystemVMessageQueue) Send(data []byte) error {
	var sysFlags int
	if mq.flags&O_NONBLOCK != 0 {
		sysFlags |= common.IpcNoWait
	}
	f := func() error { return msgsnd(mq.id, cDefaultMessageType, data, sysFlags) }
	return common.UninterruptedSyscall(f)
}

// Receive receives a message. It blocks if the queue is empty.
func (mq *SystemVMessageQueue) Receive(data []byte) error {
	var sysFlags int
	if mq.flags&O_NONBLOCK != 0 {
		sysFlags |= common.IpcNoWait
	}
	f := func() error { return msgrcv(mq.id, data, cSysVAnyMessage, sysFlags) }
	return common.UninterruptedSyscall(f)
}

// Destroy closes the queue and removes it permanently.
func (mq *SystemVMessageQueue) Destroy() error {
	if err := mq.Close(); err != nil {
		return errors.Wrap(err, "mq close failed")
	}
	err := msgctl(mq.id, common.IpcRmid, nil)
	if err == nil {
		if err = os.Remove(common.TmpFilename(mq.name)); os.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err, "failed to remove temporary file")
		}
	} else if os.IsNotExist(err) {
		err = nil
	} else {
		err = errors.Wrap(err, "msgctl failed")
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
		mq.flags &= ^O_NONBLOCK
	} else {
		mq.flags |= O_NONBLOCK
	}
	return nil
}

// DestroySystemVMessageQueue permanently removes queue with a given name.
func DestroySystemVMessageQueue(name string) error {
	mq, err := OpenSystemVMessageQueue(name, 0)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			err = nil
		} else {
			err = errors.Wrap(err, "open mq")
		}
		return err
	}
	if err = mq.Destroy(); err != nil {
		err = errors.Wrap(err, "destroy faield")
	}
	return err
}
