// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"runtime"
)

const (
	SHM_OPEN_RDONLY = 1
	SHM_OPEN_RDWR   = iota
	SHM_OPEN_CREATE
	SHM_OPEN_CREATE_IF_NOT_EXISTS
	SHM_OPEN_TRUNC
)

// MemoryRegion represents a shared memory area mapped into the address space
type MemoryRegion struct {
	impl *memoryRegionImpl
}

// Returns a new shared memory region.
// name - a name of the region. should not contain '/' and exceed 255 symbols
// mode - open mode. see SHM_OPEN* constants
// flags - a set of (probably, platform-specific) flags. see SHM_FLAG_* constants
func NewMemoryRegion(name string, size int64, mode int, flags uint32) (*MemoryRegion, error) {
	impl, err := newMemoryRegionImpl(name, size, mode, flags)
	if err != nil {
		return nil, err
	}
	result := &MemoryRegion{impl: impl}
	runtime.SetFinalizer(impl, func(object interface{}) {
		impl := object.(*memoryRegionImpl)
		impl.Close()
	})
	return result, nil
}

func (region *MemoryRegion) Destroy() error {
	return region.impl.Destroy()
}

func (region *MemoryRegion) Close() error {
	return region.impl.Close()
}
