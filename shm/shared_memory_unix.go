// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package shm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type memoryObjectImpl struct {
	file *os.File
}

func newMemoryObjectImpl(name string, mode int, perm os.FileMode) (impl *memoryObjectImpl, err error) {
	var path string
	if path, err = shmName(name); err != nil {
		return nil, err
	}
	var file *os.File
	file, err = shmOpen(path, mode, perm)
	if err != nil {
		return
	}
	impl = &memoryObjectImpl{file: file}
	return
}

// Destroy closes the object and removes it permanently
func (impl *memoryObjectImpl) Destroy() error {
	if int(impl.Fd()) >= 0 {
		if err := impl.Close(); err != nil {
			return err
		}
	}
	return destroyMemoryObject(impl.file.Name())
}

// Name returns the name of the object as it was given to NewMemoryObject()
func (impl *memoryObjectImpl) Name() string {
	result := filepath.Base(impl.file.Name())
	// on darwin we do this trick due to
	// http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if runtime.GOOS == "darwin" {
		result = result[:strings.LastIndex(result, "\t")]
	}
	return result
}

// Close closes object's underlying file object.
// Darwin: a call to Close() causes invalid argument error,
// if the object was not truncated. So, in this case we do not
// close it and return nil as an error.
func (impl *memoryObjectImpl) Close() error {
	fdBeforeClose := impl.Fd()
	err := impl.file.Close()
	if err == nil {
		return nil
	}
	if runtime.GOOS == "darwin" {
		// we're closing the file for the first time, and
		// we haven't truncated the file and it hasn't been closed
		if impl.Size() == 0 && int(fdBeforeClose) >= 0 {
			return nil
		}
	}
	return err
}

// Truncate resizes the shared memory object.
// Darwin: it is possible to truncate an object only once after it was created.
// Darwin: the size should be divisible by system page size,
// otherwise the size is set to the nearest page size devider greater, then the given size.
func (impl *memoryObjectImpl) Truncate(size int64) error {
	return impl.file.Truncate(size)
}

// Size returns the current object size.
// Darwin: it may differ from the passed passed to Truncate
func (impl *memoryObjectImpl) Size() int64 {
	fileInfo, err := impl.file.Stat()
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

// Fd returns a descriptor of the object's underlying file object
func (impl *memoryObjectImpl) Fd() uintptr {
	return impl.file.Fd()
}

// DestroyMemoryObject removes an object with a given name
func DestroyMemoryObject(name string) error {
	path, err := shmName(name)
	if err != nil {
		return err
	}
	return destroyMemoryObject(path)
}
