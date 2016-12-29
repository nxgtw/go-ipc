// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package shm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

const (
	isDarwin = runtime.GOOS == "darwin"
)

type memoryObject struct {
	file *os.File
}

func newMemoryObject(name string, flag int, perm os.FileMode) (*memoryObject, error) {
	path, err := shmName(name)
	if err != nil {
		return nil, errors.Wrap(err, "shm name failed")
	}
	file, err := shmOpen(path, flag, perm)
	if err != nil {
		return nil, errors.Wrap(err, "shm open failed")
	}
	return &memoryObject{file: file}, nil
}

func (obj *memoryObject) Destroy() error {
	if int(obj.Fd()) >= 0 {
		if err := obj.Close(); err != nil {
			return errors.Wrap(err, "close failed")
		}
	}
	if err := doDestroyMemoryObject(obj.file.Name()); err != nil {
		return errors.Wrap(err, "unable to destroy memory object")
	}
	return nil
}

func (obj *memoryObject) Name() string {
	result := filepath.Base(obj.file.Name())
	// on darwin we do this trick due to
	// http://www.opensource.apple.com/source/Libc/Libc-320/sys/shm_open.c
	if isDarwin {
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
	if isDarwin {
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
		return errors.Wrap(err, "shm name failed")
	}
	if err = doDestroyMemoryObject(path); err != nil {
		err = errors.Wrapf(err, "failed to destroy shm object %q", path)
	}
	return err
}
