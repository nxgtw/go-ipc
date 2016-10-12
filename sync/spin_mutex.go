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
	atomic.StoreUint32((*uint32)(unsafe.Pointer(spin)), cSpinUnlocked)
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
	const spinImplSize = int64(unsafe.Sizeof(spinMutex(0)))
	if !checkMutexFlags(flag) {
		return nil, errors.New("invalid open flags")
	}
	name = spinName(name)
	obj, created, resultErr := shm.NewMemoryObjectSize(name, flag, perm, spinImplSize)
	if resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm object")
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
	if region, resultErr = mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, int(spinImplSize)); resultErr != nil {
		return nil, errors.Wrap(resultErr, "failed to create shm region")
	}
	if created {
		if resultErr = allocator.Alloc(region.Data(), spinMutex(cSpinUnlocked)); resultErr != nil {
			return nil, errors.Wrap(resultErr, "failed to place mutex instance into shared memory")
		}
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

// Unlock releases the mutex.
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
