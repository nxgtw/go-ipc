// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"io"
	"os"
)

// Messenger is an interface which must be satisfied by any
// message queue implementation on any platform.
type Messenger interface {
	Send(object interface{}) error
	Receive(object interface{}) error
	io.Closer
}

func checkMqPerm(perm os.FileMode) bool {
	return uint(perm)&0111 == 0
}
