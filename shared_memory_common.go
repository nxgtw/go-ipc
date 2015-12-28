// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os"
)

// this is to ensure, that all implementations of shm-related structs
// satisfy the same minimal interface
var (
	_ iSharedMemoryObject = &MemoryObject{}
	_ iSharedMemoryRegion = &MemoryRegion{}
	_ MappableHandle      = &MemoryObject{}
)

type iSharedMemoryObject interface {
	Name() string
	Size() int64
	Truncate(size int64) error
	Close() error
	Destroy() error
}

type iSharedMemoryRegion interface {
	Data() []byte
	Size() int
	Flush(async bool) error
	Close() error
}

func calcValidOffset(offset int64) int64 {
	pageSize := int64(os.Getpagesize())
	return (offset - (offset/pageSize)*pageSize)
}

func shmCreateModeToOsMode(mode int) (int, error) {
	if mode&O_OPEN_OR_CREATE != 0 {
		if mode&(O_CREATE_ONLY|O_OPEN_ONLY) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_TRUNC | os.O_RDWR, nil
	}
	if mode&O_CREATE_ONLY != 0 {
		if mode&O_OPEN_ONLY != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_EXCL | os.O_RDWR, nil
	}
	if mode&O_OPEN_ONLY != 0 {
		return 0, nil
	}
	return 0, fmt.Errorf("no create mode flags")
}

func shmModeToOsMode(mode int) (int, error) {
	if createMode, err := shmCreateModeToOsMode(mode); err == nil {
		if accessMode, err := accessModeToOsMode(mode); err == nil {
			return createMode | accessMode, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}
