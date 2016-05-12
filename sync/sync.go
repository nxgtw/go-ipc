// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"bitbucket.org/avd/go-ipc"
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

func checkMutexOpenMode(mode int) bool {
	return mode == ipc.O_OPEN_OR_CREATE || mode == ipc.O_CREATE_ONLY || mode == ipc.O_OPEN_ONLY
}

// newMemoryObjectSize opens or creates a shared memory object with the given name.
// If the object was created, it truncates it to 'size'.
// Otherwise, checks, that the existing object is at least 'size' bytes long.
// Returns an object, true, if it was created, and an error.
func newMemoryObjectSize(name string, mode int, perm os.FileMode, size int64) (*shm.MemoryObject, bool, error) {
	var obj *shm.MemoryObject
	creator := func(create bool) error {
		var err error
		creatorMode := ipc.O_READWRITE
		if create {
			creatorMode |= ipc.O_CREATE_ONLY
		} else {
			creatorMode |= ipc.O_OPEN_ONLY
		}
		obj, err = shm.NewMemoryObject(name, creatorMode, perm)
		return err
	}
	created, resultErr := common.OpenOrCreate(creator, mode)
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
