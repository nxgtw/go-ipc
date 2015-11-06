// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

const (
	SHM_RDONLY = iota
	SMH_RDWR
	SHM_CREATE
	SSM_CREATE_ONLY
)

type MemoryRegion struct {
	impl *memoryRegionImpl
}

func NewMemoryRegion() *MemoryRegion {
	return &MemoryRegion{impl: newMemoryRegionImpl()}
}

func (region *MemoryRegion) Destroy() {

}
