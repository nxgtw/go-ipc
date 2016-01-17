// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

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

func (impl *memoryObjectImpl) Destroy() error {
	var err error
	if err = impl.Close(); err == nil {
		err = destroyMemoryObject(impl.file.Name())
	}
	return err
}

// returns the name of the object as it was given to NewMemoryObject()
func (impl *memoryObjectImpl) Name() string {
	result := filepath.Base(impl.file.Name())
	if runtime.GOOS == "darwin" {
		result = result[:strings.LastIndex(result, "\t")]
	}
	return result
}

func (impl *memoryObjectImpl) Close() error {
	return impl.file.Close()
}

func (impl *memoryObjectImpl) Truncate(size int64) error {
	return impl.file.Truncate(size)
}

func (impl *memoryObjectImpl) Size() int64 {
	fileInfo, err := impl.file.Stat()
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func (impl *memoryObjectImpl) Fd() uintptr {
	return impl.file.Fd()
}

func DestroyMemoryObject(name string) error {
	path, err := shmName(name)
	if err != nil {
		return err
	}
	return destroyMemoryObject(path)
}
