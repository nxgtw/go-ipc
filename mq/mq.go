// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package mq

import (
	"io"
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	// O_NONBLOCK flag makes mq read/write operations nonblocking.
	O_NONBLOCK = common.O_NONBLOCK
)

// Messenger is an interface which must be satisfied by any
// message queue implementation on any platform.
type Messenger interface {
	// Send sends the data. It blocks if there are no readers and the queue if full
	Send(data []byte) error
	// Receive reads data from the queue. It blocks if the queue is empty
	Receive(data []byte) error
	io.Closer
}

// TimedMessenger is a Messenger, which supports send/receive timeouts.
type TimedMessenger interface {
	Messenger
	// SendTimeout sends the data. It blocks if there are no readers and the queue if full.
	// It wait for not more, than timeout.
	SendTimeout(data []byte, timeout time.Duration) error
	// ReceiveTimeout reads data from the queue. It blocks if the queue is empty.
	// It wait for not more, than timeout.
	ReceiveTimeout(data []byte, timeout time.Duration) error
}

// New creates a mq with a given name and permissions.
// It uses the default implementation. If there are several implementations on a platform,
// you can use explicit create functions.
//	name - unique queue name.
//	perm - permissions for the new queue. this may not be supported by all implementations.
func New(name string, perm os.FileMode) (Messenger, error) {
	return createMQ(name, perm)
}

// Open opens a mq with a given name and flags.
// It uses the default implementation. If there are several implementations on a platform,
// you can use explicit create functions.
//	name  - unique queue name.
//	flags - a set of flags can be used to specify r/w options. this may not be supported by all implementations.
func Open(name string, flags int) (Messenger, error) {
	return openMQ(name, flags)
}

// Destroy permanently removes mq object.
func Destroy(name string) error {
	return destroyMq(name)
}

func checkMqPerm(perm os.FileMode) bool {
	return uint(perm)&0111 == 0
}
