// Copyright 2015 Aleksandr Demakin. All rights reserved.

package shm

import (
	"os"
	"path/filepath"
	"runtime"
)

// Shared memory on Windows is emulated via regular files
// like it is done in boost c++ library
type memoryObjectImpl struct {
	file *os.File
}

func newMemoryObjectImpl(name string, mode int, perm os.FileMode) (impl *memoryObjectImpl, err error) {
	path, err := shmName(name)
	if err != nil {
		return nil, err
	}
	osMode, err := openModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, osMode, perm)
	if err != nil {
		return nil, err
	}
	return &memoryObjectImpl{file}, nil
}

func (impl *memoryObjectImpl) Destroy() error {
	if int(impl.Fd()) >= 0 {
		if err := impl.Close(); err != nil {
			return err
		}
	}
	return DestroyMemoryObject(impl.Name())
}

func (impl *memoryObjectImpl) Name() string {
	return filepath.Base(impl.file.Name())
}

func (impl *memoryObjectImpl) Close() error {
	runtime.SetFinalizer(impl, nil)
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

// DestroyMemoryObject permanently removes given memory object
func DestroyMemoryObject(name string) error {
	path, err := shmName(name)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func shmName(name string) (string, error) {
	path, err := sharedDirName()
	if err != nil {
		return "", err
	}
	return path + "/" + name, nil
}

func sharedDirName() (string, error) {
	rootPath := os.TempDir() + "/go-ipc"
	if err := os.Mkdir(rootPath, 0644); err != nil && !os.IsExist(err) {
		return "", err
	}
	return rootPath, nil
}
