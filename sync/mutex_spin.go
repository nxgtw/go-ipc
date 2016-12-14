// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"runtime"
	"time"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
)

const (
	cSpinUnlocked = 0
	cSpinLocked   = 1
)

// all implementations must satisfy IPCLocker interface.
var (
	_ IPCLocker = (*SpinMutex)(nil)
)

// SpinMutex is a synchronization object which performs busy wait loop.
type SpinMutex struct {
	lwm    *lwMutex
	region *mmf.MemoryRegion
	name   string
}

type spinWW struct{}

func (sw spinWW) wake(uint32) (int, error) {
	return 1, nil
}

func (sw spinWW) wait(unused uint32, timeout time.Duration) error {
	runtime.Gosched()
	return nil
}

// NewSpinMutex creates a new spin mutex.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewSpinMutex(name string, flag int, perm os.FileMode) (*SpinMutex, error) {
	if err := ensureOpenFlags(flag); err != nil {
		return nil, err
	}
	name = spinName(name)
	region, created, err := createWritableRegion(name, flag, perm, lwmCellSize, nil)
	if err != nil {
		return nil, err
	}
	lwm := newLightweightMutex(allocator.ByteSliceData(region.Data()), &spinWW{})
	if created {
		lwm.init()
	}
	impl := &SpinMutex{region: region, name: name, lwm: lwm}
	return impl, nil
}

// Lock locks the mutex waiting in a busy loop if needed.
func (spin *SpinMutex) Lock() {
	spin.lwm.lock()
}

// LockTimeout locks the mutex waiting in a busy loop for not longer, than timeout.
func (spin *SpinMutex) LockTimeout(timeout time.Duration) bool {
	return spin.lwm.lockTimeout(timeout)
}

// Unlock releases the mutex. It panics, if the mutex is not locked.
func (spin *SpinMutex) Unlock() {
	spin.lwm.unlock()
}

// TryLock makes one attempt to lock the mutex. It return true on succeess and false otherwise.
func (spin *SpinMutex) TryLock() bool {
	return spin.lwm.tryLock()
}

// Close indicates, that the object is no longer in use,
// and that the underlying resources can be freed.
func (spin *SpinMutex) Close() error {
	return spin.region.Close()
}

// Destroy removes the mutex object.
func (spin *SpinMutex) Destroy() error {
	if err := spin.Close(); err != nil {
		return errors.Wrap(err, "failed to close spin mutex")
	}
	spin.region = nil
	err := shm.DestroyMemoryObject(spin.name)
	spin.name = ""
	if err != nil {
		return errors.Wrap(err, "failed to destroy shm object")
	}
	return nil
}

// DestroySpinMutex removes a mutex object with the given name
func DestroySpinMutex(name string) error {
	return shm.DestroyMemoryObject(spinName(name))
}

func spinName(name string) string {
	return "go-ipc.spin." + name
}
