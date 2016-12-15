// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"io"
	"os"
	"sync"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
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

// ensureOpenFlags ensures, that no other flags but os.O_CREATE and os.O_EXCL are set.
func ensureOpenFlags(flags int) error {
	if flags & ^(os.O_CREATE|os.O_EXCL) != 0 {
		return errors.New("only os.O_CREATE and os.O_EXCL are allowed")
	}
	return nil
}

func createWritableRegion(name string, flag int, perm os.FileMode, size int, init interface{}) (*mmf.MemoryRegion, bool, error) {
	obj, created, resultErr := shm.NewMemoryObjectSize(name, flag, perm, int64(size))
	if resultErr != nil {
		return nil, false, errors.Wrap(resultErr, "failed to create shm object")
	}
	var region *mmf.MemoryRegion
	defer func() {
		obj.Close()
		if resultErr == nil {
			return
		}
		if region != nil {
			region.Close()
		}
		if created {
			obj.Destroy()
		}
	}()
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, size); resultErr != nil {
		return nil, false, errors.Wrap(resultErr, "failed to create shm region")
	}
	if created && init != nil {
		if resultErr = allocator.Alloc(region.Data(), init); resultErr != nil {
			return nil, false, errors.Wrap(resultErr, "failed to place an object into shared memory")
		}
	}
	return region, created, nil
}
