// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"bytes"
	"errors"
	"io"
	"os"
	"runtime"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
)

// MemoryRegion is a mmapped area of a memory object.
// Warning. The internal object has a finalizer set,
// so the region will be unmapped during the gc.
// Thus, you should be carefull getting internal data.
// For example, the following code may crash:
// func f() {
// 	region := NewMemoryRegion(...)
// 	return g(region.Data())
// region may be gc'ed while its data is used by g()
// To avoid this, you can use UseMemoryRegion() or region readers/writers
type MemoryRegion struct {
	*memoryRegion
}

// MappableHandle is an object, which can return a handle,
// that can be used as a file descriptor for mmap
type MappableHandle interface {
	Fd() uintptr
}

// MemoryRegionReader is a reader for safe operations over a shared memory region.
// It holds a reference to the region, so the former can't be gc'ed
type MemoryRegionReader struct {
	region *MemoryRegion
	*bytes.Reader
}

// NewMemoryRegionReader creates a new reader for the given region
func NewMemoryRegionReader(region *MemoryRegion) *MemoryRegionReader {
	return &MemoryRegionReader{
		region: region,
		Reader: bytes.NewReader(region.Data()),
	}
}

// MemoryRegionWriter is a writer for safe operations over a shared memory region.
// It holds a reference to the region, so the former can't be gc'ed
type MemoryRegionWriter struct {
	region *MemoryRegion
}

// NewMemoryRegionWriter creates a new writer for the given region
func NewMemoryRegionWriter(region *MemoryRegion) *MemoryRegionWriter {
	return &MemoryRegionWriter{region: region}
}

// WriteAt is to implement io.WriterAt
func (w *MemoryRegionWriter) WriteAt(p []byte, off int64) (n int, err error) {
	data := w.region.Data()
	n = len(data) - int(off)
	if n > 0 {
		if n > len(p) {
			n = len(p)
		}
		copy(data[off:], p[:n])
	}
	if n < len(p) {
		err = io.EOF
	}
	return
}

// NewMemoryRegion creates a new shared memory region.
// object - an object containing a descriptor of the file, which can be mmaped
// size - object size
// mode - open mode. see MEM_* constants
// offset - offset in bytes from the beginning of the mmaped file
// size - region size
func NewMemoryRegion(object MappableHandle, mode int, offset int64, size int) (*MemoryRegion, error) {
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
func UseMemoryRegion(region *MemoryRegion) {
	allocator.Use(unsafe.Pointer(region))
}

// calcMmapOffsetFixup returns a value X,
// so that  offset - X is a valid mmap offset
// typically the value of the fixup is a memory page size,
// however, on windows it must be a multiple of the
// memory allocation granularity value as well.
func calcMmapOffsetFixup(offset int64) int64 {
	pageSize := mmapOffsetMultiple()
	return (offset - (offset/pageSize)*pageSize)
}

// fileInfoGetter is used to obtain size of the object
type fileInfoGetter interface {
	Stat() (os.FileInfo, error)
}

func fileSizeFromFd(f MappableHandle) (int64, error) {
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

func checkMmapSize(f MappableHandle, size int) (int, error) {
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
