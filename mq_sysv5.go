// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"fmt"
	"os"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

const (
	cIpcCreate = 00001000 /* create if key is nonexistent */
	cIpcExcl   = 00002000 /* fail if key exists */
	cIpcNoWait = 00004000 /* return error on wait */

	cIpcRmid = 0 /* remove resource */
	cIpcSet  = 1 /* set ipc_perm options */
	cIpcStat = 2 /* get ipc_perm options */
	cIpcInfo = 3 /* see ipcs */

	cDefaultMessageType = 1
)

// SystemVMessageQueue is a System V ipc mechanism based on message passing
type SystemVMessageQueue struct {
	flags int
	id    int
	name  string
}

type key uint64

// this is to ensure, that system V implementation of ipc mq
// satisfy the minimal queue interface
var (
	_ Messenger = (*SystemVMessageQueue)(nil)
)

// CreateSystemVMessageQueue creates new queue with a given name and permissions.
// 'x' permission cannot be used.
func CreateSystemVMessageQueue(name string, perm os.FileMode) (*SystemVMessageQueue, error) {
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	k, err := sysVMqkey(name)
	if err != nil {
		return nil, err
	}
	id, err := msgget(k, int(perm)|cIpcCreate|cIpcExcl)
	if err != nil {
		return nil, err
	}
	return &SystemVMessageQueue{id: id, name: name}, nil
}

// OpenSystemVMessageQueue opens existing message queue
func OpenSystemVMessageQueue(name string, flags int) (*SystemVMessageQueue, error) {
	k, err := sysVMqkey(name)
	if err != nil {
		return nil, err
	}
	id, err := msgget(k, 0)
	if err != nil {
		return nil, err
	}
	result := &SystemVMessageQueue{id: id, name: name}
	if flags&O_NONBLOCK != 0 {
		result.flags |= cIpcNoWait
	}
	return result, nil
}

// Send sends a message.
// It blocks if the queue is full.
func (mq *SystemVMessageQueue) Send(object interface{}) error {
	data, err := allocator.ObjectData(object)
	if err != nil {
		return err
	}
	return msgsnd(mq.id, cDefaultMessageType, data, mq.flags)
}

// Receive receives a message.
// It blocks if the queue is empty.
func (mq *SystemVMessageQueue) Receive(object interface{}) error {
	if !allocator.IsReferenceType(object) {
		return fmt.Errorf("expected a slice, or a pointer")
	}
	data, err := allocator.ObjectData(object)
	if err != nil {
		return err
	}
	// 0 - receive any messages
	return msgrcv(mq.id, data, 0, mq.flags)
}

// Destroy closes the queue and removes it permanently
func (mq *SystemVMessageQueue) Destroy() error {
	mq.Close()
	err := msgctl(mq.id, cIpcRmid, nil)
	if err == nil {
		if err = os.Remove(filenameForSysVMqName(mq.name)); os.IsNotExist(err) {
			err = nil
		}
	} else if os.IsNotExist(err) {
		err = nil
	}
	return err
}

// Close closes the queue. There is no need to close System V mq.
func (mq *SystemVMessageQueue) Close() error {
	return nil
}

// SetBlocking sets whether the send/receive operations on the queue block.
func (mq *SystemVMessageQueue) SetBlocking(block bool) error {
	if block {
		mq.flags &= ^cIpcNoWait
	} else {
		mq.flags |= cIpcNoWait
	}
	return nil
}

// DestroySystemVMessageQueue permanently removes queue with a given name
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

func sysVMqkey(name string) (key, error) {
	name = filenameForSysVMqName(name)
	file, err := os.Create(name)
	if err != nil {
		return 0, errors.New("invalid mq name")
	}
	file.Close()
	k, err := ftok(name)
	if err != nil {
		return 0, errors.New("invalid mq name")
	}
	return k, nil
}

func filenameForSysVMqName(name string) string {
	return os.TempDir() + "/" + name
}

func ftok(name string) (key, error) {
	var statfs unix.Stat_t
	if err := unix.Stat(name, &statfs); err != nil {
		return key(0), err
	}
	return key(statfs.Ino&0xFFFF | ((statfs.Dev & 0xFF) << 16)), nil
}
