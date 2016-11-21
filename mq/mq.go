// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

import (
	"io"
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	// O_NONBLOCK flag makes mq send/receive operations non-blocking.
	O_NONBLOCK = common.O_NONBLOCK
)

// Blocker is an object, which can work in blocking and non-blocking modes.
type Blocker interface {
	SetBlocking(bool) error
}

// Buffered is an object with internal buffer of the given capacity.
type Buffered interface {
	Cap() (int, error)
}

// Messenger is an interface which must be satisfied by any
// message queue implementation on any platform.
type Messenger interface {
	// Send sends the data. It blocks if there are no readers and the queue is full
	Send(data []byte) error
	// Receive reads data from the queue. It blocks if the queue is empty.
	// Returns message len.
	Receive(data []byte) (int, error)
	io.Closer
}

// TimedMessenger is a Messenger, which supports send/receive timeouts.
// Passing 0 as a timeout makes a call non-blocking.
// Passing negative value as a timeout makes the timeout infinite.
type TimedMessenger interface {
	Messenger
	// SendTimeout sends the data. It blocks if there are no readers and the queue if full.
	// It waits for not more, than timeout.
	SendTimeout(data []byte, timeout time.Duration) error
	// ReceiveTimeout reads data from the queue. It blocks if the queue is empty.
	// It waits for not more, than timeout. Returns message len.
	ReceiveTimeout(data []byte, timeout time.Duration) (int, error)
}

// PriorityMessenger is a Messenger, which orders messages according to their priority.
// Semantic is similar to linux native mq:
// Messages are placed on the queue in decreasing order of priority, with newer messages of the same
// priority being placed after older messages with the same priority.
type PriorityMessenger interface {
	Messenger
	Buffered
	// SendPriority sends the data. The message will be inserted in the mq according to its priority.
	SendPriority(data []byte, prio int) error
	// ReceivePriority reads a message and returns its len and priority.
	ReceivePriority(data []byte) (int, int, error)
}

// New creates a mq with a given name and permissions.
// It uses the default implementation. If there are several implementations on a platform,
// you can use explicit create functions.
//	name - unique queue name.
//	flag - create flags. You can specify:
//		os.O_EXCL if you don't want to open a queue if it exists.
//		O_NONBLOCK if you don't want to block on send/receive.
//			This flag may not be supported by a particular implementation. To be sure, you can convert Messenger
//			to Blocker and call SetBlocking to set/unset non-blocking mode.
//	perm - permissions for the new queue.
func New(name string, flag int, perm os.FileMode) (Messenger, error) {
	return createMQ(name, flag, perm)
}

// Open opens a mq with a given name and flags.
// It uses the default implementation. If there are several implementations on a platform,
// you can use explicit create functions.
//	name  - unique queue name.
//	flags - 0 or O_NONBLOCK.
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

// IsTemporary returns true, if an error is a timeout error.
func IsTemporary(err error) bool {
	return common.IsTimeoutErr(err) || isTemporaryError(err)
}
