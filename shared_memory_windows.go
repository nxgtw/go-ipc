// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"path/filepath"
	"os"
	"fmt"
	
	"golang.org/x/sys/windows"
)

type memoryObjectImpl struct {
	file *os.File
}

// Shared memory on Windows is emulated via usual files
// like it is done in boost c++ library
type memoryRegionImpl struct {
	data       []byte
	size       int
	pageOffset int64
}

func newMemoryObjectImpl(name string, mode int, perm os.FileMode) (impl *memoryObjectImpl, err error) {
	path,  err := shmName(name)
	if err != nil {
		return nil, err
	}
	osMode, err := modeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, osMode, perm)
	if err != nil {
		return nil, err
	}
	return &memoryObjectImpl{file},  nil
}

func (impl *memoryObjectImpl) Destroy() error {
	if err := impl.Close(); err == nil {
		return os.Remove(impl.file.Name())
	} else {
		return err
	}
}

func (impl *memoryObjectImpl) Name() string {
	return filepath.Base(impl.file.Name())
}

func (impl *memoryObjectImpl) Close() error {
	return impl.file.Close()
}

func (impl *memoryObjectImpl) Truncate(size int64) error {
	return impl.file.Truncate(size)
}

func (impl *memoryObjectImpl) Size() int64 {
	if fileInfo, err := impl.file.Stat(); err != nil {
		return 0
	} else {
		return fileInfo.Size()
	}
}

func (impl *memoryObjectImpl) Fd() int {
	return int(impl.file.Fd())
}

func newMemoryRegionImpl(obj MappableHandle, mode int, offset int64, size int) (*memoryRegionImpl, error) {
	prot, flags, err := shmProtAndFlagsFromMode(mode)
	if err != nil {
		return nil, err
	}
	//file:///home/avd/dev/boost_1_59_0/boost/interprocess/mapped_region.hpp:441
	// TODO(avd) - check if it is not for shm
	handle := windows.InvalidHandle
	if true {
		// TODO(avd) - security attrrs
		var err error
		if handle, err = windows.CreateFileMapping(windows.Handle(obj.Fd()), nil, prot , 0, 0, nil); err != nil {
			return nil, err
		}
	} else {		
		// TODO(avd) - finish with it
	}
	if size == 0 { // TODO(avd) get current file size
		
	}
	defer windows.CloseHandle(handle)
	pageOffset := calcValidOffset(offset)
	lowOffset := uint32(pageOffset)
	highOffset := uint32(pageOffset>> 32)
	addr, err := windows.MapViewOfFile(handle, flags, lowOffset, highOffset, uintptr(int64(size) + pageOffset))
	if err != nil {
		return nil, err
	}
	sz := size + int(pageOffset)
	return &memoryRegionImpl{byteSliceFromUntptr(addr,sz, sz), size, pageOffset}, nil
}

func (impl *memoryRegionImpl) Close() error {
	return windows.UnmapViewOfFile(byteSliceAddress(impl.data))
}

func (impl *memoryRegionImpl) Data() []byte {
	return impl.data[impl.pageOffset:]
}

func (impl *memoryRegionImpl) Size() int {
	return impl.size
}

func (impl *memoryRegionImpl) Flush(async bool) error {
	return windows.FlushViewOfFile(byteSliceAddress(impl.data), uintptr(len(impl.data)))
}

func DestroyMemoryObject(name string) error {
	if path, err := shmName(name); err != nil {
		return err
	} else {
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
}

func shmProtAndFlagsFromMode(mode int) (prot uint32, flags uint32, err error) {
	switch mode {
	case SHM_READ_ONLY:
		fallthrough
	case SHM_READ_PRIVATE:
		prot = windows.PAGE_READONLY
		flags = windows.FILE_MAP_READ
	case SHM_READWRITE:
		prot = windows.PAGE_READWRITE
		flags = windows.FILE_MAP_WRITE
	case SHM_COPY_ON_WRITE:
		prot = windows.PAGE_WRITECOPY
		flags = windows.FILE_MAP_COPY
	default:
		err = fmt.Errorf("invalid shm region flags")
	}
	return
}

func shmName(name string) (string, error) {
	if path, err  := sharedDirName(); err != nil {
		return "", err
	} else {
		return path + "/" + name, nil
	}
}

func sharedDirName() (string,  error) {
	rootPath :=  os.TempDir() + "/go-ipc"
	if err := os.Mkdir(rootPath, 0644); err == nil || os.IsExist(err) {
		return rootPath, nil
	} else {
		return "", err
	}
}