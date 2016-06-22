// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"io"
	"os"
	"sync"
	"time"
)

// IPCLocker is a minimal interface, which must be satisfied by any synchronization primitive
// on any platform.
type IPCLocker interface {
	sync.Locker
	io.Closer
}

// TimedIPCLocker is a locker, whose lock operation can be limited with duration.
type TimedIPCLocker interface {
	IPCLocker
	// LockTimeout tries to lock the locker, waiting for not more, than timeout
	LockTimeout(timeout time.Duration) bool
}

func checkMutexFlags(flags int) bool {
	return flags & ^(os.O_CREATE|os.O_EXCL) == 0
}
