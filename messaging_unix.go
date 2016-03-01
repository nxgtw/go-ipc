// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

const (
	cIpcCreate = 00001000
	cIpcExcl   = 00002000
	cIpcNoWait = 00004000
)

// MessageQueue is a unix-specific ipc mechanism based on message passing
type MessageQueue struct {
	id int
}

type key uint64

// CreateMessageQueue creates new queue with a given name and permissions.
// 'x' permission cannot be used.
func CreateMessageQueue(name string, perm os.FileMode) (*MessageQueue, error) {
	if !checkMqPerm(perm) {
		return nil, errors.New("invalid mq permissions")
	}
	k, err := keyForMq(name)
	if err != nil {
		return nil, err
	}
	id, err := msgget(k, int(perm)|cIpcCreate|cIpcExcl)
	if err != nil {
		return nil, err
	}
	return &MessageQueue{id: id}, nil
}

// OpenMessageQueue opens existing message queue
func OpenMessageQueue(name string) (*MessageQueue, error) {
	k, err := keyForMq(name)
	if err != nil {
		return nil, err
	}
	id, err := msgget(k, 0)
	if err != nil {
		return nil, err
	}
	return &MessageQueue{id: id}, nil
}

// Send sends a message.
// It blocks if the queue is full.
func (mq *MessageQueue) Send(object interface{}) error {
	data, err := objectByteSlice(object)
	if err != nil {
		return err
	}
	return msgsnd(mq.id, 0, data, 0)
}

// Receive receives a message.
// It blocks if the queue is empty.
func (mq *MessageQueue) Receive(object interface{}) error {
	//	return mq.ReceiveTimeout(object, prio, time.Duration(-1))
	return nil
}

func keyForMq(name string) (key, error) {
	name = filenameForMqName(name)
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

func filenameForMqName(name string) string {
	return os.TempDir() + "/" + name
}

func ftok(name string) (key, error) {
	var statfs unix.Stat_t
	if err := unix.Stat(name, &statfs); err != nil {
		return key(0), err
	}
	return key(statfs.Ino&0xFFFF | ((statfs.Dev & 0xFF) << 16)), nil
}
