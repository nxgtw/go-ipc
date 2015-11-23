// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"runtime"
)

const (
	SHM_RDONLY = iota
	SMH_RDWR
	SHM_CREATE
	SHM_CREATE_ONLY
)

type MemoryRegion struct {
	impl *memoryRegionImpl
}

func NewMemoryRegion(name string, size uint64, mode int, flags uint32) (*MemoryRegion, error) {
	result := &MemoryRegion{impl: newMemoryRegionImpl()}
	runtime.SetFinalizer(result, func(object interface{}) {
		//TODO (avd) - define destroy behavior
	})
	return result, nil
}

func (region *MemoryRegion) Destroy() {

}
