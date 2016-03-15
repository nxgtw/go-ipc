// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"io"
	"os"
	"time"
)

// Messenger is an interface which must be satisfied by any
// message queue implementation on any platform.
type Messenger interface {
	Send(object interface{}) error
	Receive(object interface{}) error
	io.Closer
}

// TimedMessenger is a Messenger, which supports send/receive timeouts.
type TimedMessenger interface {
	Messenger
	SendTimeout(object interface{}, timeout time.Duration) error
	ReceiveTimeout(object interface{}, timeout time.Duration) error
}

func checkMqPerm(perm os.FileMode) bool {
	return uint(perm)&0111 == 0
}

// CreateMQ creates a mq with a given name and permissions.
// It uses the default implementation. If there are several implementations on a platform,
// you should use explicit create functions.
func CreateMQ(name string, perm os.FileMode) (Messenger, error) {
	return createMQ(name, perm)
}

// OpenMQ opens a mq with a given name and flags.
// It uses the default implementation. If there are several implementations on a platform,
// you should use explicit create functions.
func OpenMQ(name string, flags int) (Messenger, error) {
	return openMQ(name, flags)
}

func DestroyMQ(name string) error {
	return destroyMq(name)
}
