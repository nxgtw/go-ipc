// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

const (
	SHM_RDONLY = iota
	SMH_RDWR
	SHM_CREATE
	SHM_CREATE_ONLY
)

type MemoryRegion struct {
	impl *memoryRegionImpl
}

func NewMemoryRegion(name string, size uint64) *MemoryRegion {
	return &MemoryRegion{impl: newMemoryRegionImpl()}
}

func (region *MemoryRegion) Destroy() {

}
