// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
)

type memoryObjectImpl struct {
	file *os.File
}

// Shared memory on Windows is emulated via usual files
// like it is done in boost c++ library
type memoryRegionImpl struct {
	data       []byte
	size       int
}

func newMemoryObjectImpl(name string, mode int, perm os.FileMode) (impl *memoryObjectImpl, err error) {
	fullName, err := shmName(name)
	if err != nil {
		return nil, err
	}
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