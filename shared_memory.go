// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"runtime"
)

const (
	SHM_RDONLY = iota
	SHM_RDWR
	SHM_CREATE
	SHM_CREATE_IF_NOT_EXISTS
	SHM_TRUNC
)

type MemoryRegion struct {
	impl *memoryRegionImpl
}

func NewMemoryRegion(name string, size int64, mode int, flags uint32) (*MemoryRegion, error) {
	impl, err := newMemoryRegionImpl(name, size, mode, flags)
	if err != nil {
		return nil, err
	}
	result := &MemoryRegion{impl: impl}
	runtime.SetFinalizer(result, func(object interface{}) {
		//TODO (avd) - define destroy behavior
	})
	return result, nil
}

func (region *MemoryRegion) Destroy() {
	region.impl.Destroy()
}

func (region *MemoryRegion) Close() {
	region.impl.Close()
}
