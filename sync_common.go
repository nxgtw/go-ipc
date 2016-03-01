// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"sync"
)

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*SpinMutex)(nil)
	_ IPCLocker = (*Mutex)(nil)
)

// IPCLocker is a minimal interface, which must be satisfied by any synchronization primitive
// on any platform
type IPCLocker interface {
	sync.Locker
	Finish() error
}

func checkMutexOpenMode(mode int) bool {
	return mode == O_OPEN_OR_CREATE || mode == O_CREATE_ONLY || mode == O_OPEN_ONLY
}
