// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"errors"
	"os"
	"runtime"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
)

var (
	mmapOffsetMultiple int64
)

// MemoryRegion is a mmapped area of a memory object.
// Warning. The internal object has a finalizer set,
// so the region will be unmapped during the gc.
// Thus, you should be carefull getting internal data.
// For example, the following code may crash:
// 	func f() {
// 		region := NewMemoryRegion(...)
// 		return g(region.Data())
// 	}
// region may be gc'ed while its data is used by g().
// To avoid this, you can use UseMemoryRegion() or region readers/writers.
type MemoryRegion struct {
	*memoryRegion
}

// Mappable is a named object, which can return a handle,
// that can be used as a file descriptor for mmap.
type Mappable interface {
	Fd() uintptr
	Name() string
}

// NewMemoryRegion creates a new shared memory region.
// 	object - an object to mmap.
// 	mode - open mode. see MEM_* constants
// 	offset - offset in bytes from the beginning of the mmaped file
// 	size - mapping size.
func NewMemoryRegion(object Mappable, mode int, offset int64, size int) (*MemoryRegion, error) {
	impl, err := newMemoryRegion(object, mode, offset, size)
	if err != nil {
		return nil, err
	}
	result := &MemoryRegion{impl}
	runtime.SetFinalizer(impl, func(region *memoryRegion) {
		region.Close()
	})
	return result, nil
}

// Close unmaps the regions so that it cannot be longer used.
func (region *MemoryRegion) Close() error {
	return region.memoryRegion.Close()
}

// Data returns region's mapped data.
// This function can be dangerous and could be removed in future releases.
func (region *MemoryRegion) Data() []byte {
	return region.memoryRegion.Data()
}

// Flush syncs mapped content with the file data.
func (region *MemoryRegion) Flush(async bool) error {
	return region.memoryRegion.Flush(async)
}

// Size returns mapping size.
func (region *MemoryRegion) Size() int {
	return region.memoryRegion.Size()
}

// UseMemoryRegion ensures, that the object is still alive at the moment of the call.
// The usecase is when you use memory region's Data() and don't use the
// region itself anymore. In this case the region can be gc'ed, the memory mapping
// destroyed and you can get segfault.
// It can be used like the following:
// 	region := NewMemoryRegion(...)
//	defer UseMemoryRegion(region)
// 	data := region.Data()
//	{ work with data }
// However, it is better to use MemoryRegionReader/Writer.
// This function could be removed in future releases.
func UseMemoryRegion(region *MemoryRegion) {
	allocator.Use(unsafe.Pointer(region))
}

// calcMmapOffsetFixup returns a value X,
// so that  offset - X is a valid mmap offset
// typically the value of the fixup is a memory page size,
// however, on windows it must be a multiple of the
// memory allocation granularity value as well.
func calcMmapOffsetFixup(offset int64) int64 {
	return (offset - (offset/mmapOffsetMultiple)*mmapOffsetMultiple)
}

// fileInfoGetter is used to obtain file's size
type fileInfoGetter interface {
	Stat() (os.FileInfo, error)
}

func fileSizeFromFd(f Mappable) (int64, error) {
	if f.Fd() == ^uintptr(0) {
		return 0, nil
	}
	if ig, ok := f.(fileInfoGetter); ok {
		fi, err := ig.Stat()
		if err != nil {
			return 0, err
		}
		return fi.Size(), nil
	}
	return 0, nil
}

func checkMmapSize(f Mappable, size int) (int, error) {
	if size == 0 {
		if f.Fd() == ^uintptr(0) {
			return 0, errors.New("must provide a valid file size")
		}
		if sz, err := fileSizeFromFd(f); err == nil {
			size = int(sz)
		} else {
			return 0, err
		}
	}
	return size, nil
}
