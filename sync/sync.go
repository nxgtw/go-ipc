// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"io"
	"sync"
	"time"

	"bitbucket.org/avd/go-ipc"
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
	io.Closer
}

// IPCLocker is a locker, whose lock operation can be limited with duration
type TimedIPCLocker interface {
	IPCLocker
	LockTimeout(timeout time.Duration) bool
}

func checkMutexOpenMode(mode int) bool {
	return mode == ipc.O_OPEN_OR_CREATE || mode == ipc.O_CREATE_ONLY || mode == ipc.O_OPEN_ONLY
}
