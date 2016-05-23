// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
	"bitbucket.org/avd/go-ipc/shm"
)

// this is to ensure, that all implementations of ipc mutex
// satisfy the same minimal interface
var (
	_ IPCLocker = (*SpinMutex)(nil)
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

// newMemoryObjectSize opens or creates a shared memory object with the given name.
// If the object was created, it truncates it to 'size'.
// Otherwise, checks, that the existing object is at least 'size' bytes long.
// Returns an object, true, if it was created, and an error.
func newMemoryObjectSize(name string, flag int, perm os.FileMode, size int64) (*shm.MemoryObject, bool, error) {
	var obj *shm.MemoryObject
	creator := func(create bool) error {
		var err error
		creatorFlag := os.O_RDWR
		if create {
			creatorFlag |= (os.O_CREATE | os.O_EXCL)
		}
		obj, err = shm.NewMemoryObject(name, creatorFlag, perm)
		return err
	}
	created, resultErr := common.OpenOrCreate(creator, flag)
	if resultErr != nil {
		return nil, false, resultErr
	}
	if created {
		if resultErr = obj.Truncate(size); resultErr != nil {
			return nil, false, resultErr
		}
	} else if obj.Size() < size {
		return nil, false, fmt.Errorf("existing object is not big enough")
	}
	return obj, created, nil
}
