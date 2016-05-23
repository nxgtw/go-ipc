// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package shm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type memoryObject struct {
	file *os.File
}

func newMemoryObject(name string, flag int, perm os.FileMode) (impl *memoryObject, err error) {
	var path string
	if path, err = shmName(name); err != nil {
		return nil, err
	}
	var file *os.File
	file, err = shmOpen(path, flag, perm)
	if err != nil {
		return
	}
	impl = &memoryObject{file: file}
	return
}

func (obj *memoryObject) Destroy() error {
	if int(obj.Fd()) >= 0 {
		if err := obj.Close(); err != nil {
			return err
		}
	}
	return doDestroyMemoryObject(obj.file.Name())
}

func (obj *memoryObject) Name() string {
	result := filepath.Base(obj.file.Name())
	// on darwin we do this trick due to
	// http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if runtime.GOOS == "darwin" {
		result = result[:strings.LastIndex(result, "\t")]
	}
	return result
}

func (obj *memoryObject) Close() error {
	fdBeforeClose := obj.Fd()
	err := obj.file.Close()
	if err == nil {
		return nil
	}
	if runtime.GOOS == "darwin" {
		// we're closing the file for the first time, and
		// we haven't truncated the file and it hasn't been closed
		if obj.Size() == 0 && int(fdBeforeClose) >= 0 {
			return nil
		}
	}
	return err
}

func (obj *memoryObject) Truncate(size int64) error {
	return obj.file.Truncate(size)
}

func (obj *memoryObject) Size() int64 {
	fileInfo, err := obj.file.Stat()
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func (obj *memoryObject) Fd() uintptr {
	return obj.file.Fd()
}

func destroyMemoryObject(name string) error {
	path, err := shmName(name)
	if err != nil {
		return err
	}
	return doDestroyMemoryObject(path)
}
