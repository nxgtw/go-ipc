// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"io"
	"os"
	"sync"
	"time"
)

// IPCLocker is a minimal interface, which must be satisfied by any synchronization primitive on any platform.
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

// NewMutex creates a new interprocess mutex.
// It uses the default implementation on the current platform.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewMutex(name string, flag int, perm os.FileMode) (TimedIPCLocker, error) {
	return newMutex(name, flag, perm)
}

// DestroyMutex permanently removes mutex with the given name.
func DestroyMutex(name string) error {
	return destroyMutex(name)
}

func mutexSharedStateName(name, typ string) string {
	return name + ".s" + typ
}
