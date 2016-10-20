// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"

	"github.com/pkg/errors"
)

const (
	spinImplSize  = int(unsafe.Sizeof(spinMutex(0)))
	cSpinUnlocked = 0
	cSpinLocked   = 1
)

// all implementations must satisfy IPCLocker interface.
var (
	_ IPCLocker = (*SpinMutex)(nil)
)

type spinMutex uint32

func (spin *spinMutex) lock() {
	for !spin.tryLock() {
		runtime.Gosched()
	}
}

func (spin *spinMutex) lockTimeout(timeout time.Duration) bool {
	var attempt uint64
	start := time.Now()
	for !spin.tryLock() {
		runtime.Gosched()
		if attempt%1000 == 0 { // do not call time.Since too often.
			if timeout >= 0 && time.Since(start) >= timeout {
				return false
			}
		}
		attempt++
	}
	return true
}

func (spin *spinMutex) unlock() {
	if !atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(spin)), cSpinLocked, cSpinUnlocked) {
		panic("unlock of unlocked mutex")
	}
}

func (spin *spinMutex) tryLock() bool {
	return atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(spin)), cSpinUnlocked, cSpinLocked)
}

// SpinMutex is a synchronization object which performs busy wait loop.
type SpinMutex struct {
	impl   *spinMutex
	region *mmf.MemoryRegion
	name   string
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
	region, err := createWritableRegion(name, flag, perm, spinImplSize, spinMutex(cSpinUnlocked))
	if err != nil {
		return nil, err
	}
	m := (*spinMutex)(allocator.ByteSliceData(region.Data()))
	impl := &SpinMutex{impl: m, region: region, name: name}
	return impl, nil
}

// Lock locks the mutex waiting in a busy loop if needed.
func (spin *SpinMutex) Lock() {
	spin.impl.lock()
}

// LockTimeout locks the mutex waiting in a busy loop for not longer, than timeout.
func (spin *SpinMutex) LockTimeout(timeout time.Duration) bool {
	return spin.impl.lockTimeout(timeout)
}

// Unlock releases the mutex. It panics, if the mutex is not locked.
func (spin *SpinMutex) Unlock() {
	spin.impl.unlock()
}

// TryLock makes one attempt to lock the mutex. It return true on succeess and false otherwise.
func (spin *SpinMutex) TryLock() bool {
	return spin.impl.tryLock()
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
