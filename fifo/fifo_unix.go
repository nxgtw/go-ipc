// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package fifo

import (
	"fmt"
	"os"
	"strings"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

type fifoImpl struct {
	file *os.File
}

func newFifoImpl(name string, mode int, perm os.FileMode) (*fifoImpl, error) {
	if _, err := common.CreateModeToOsMode(mode); err != nil {
		return nil, err
	}
	osMode, err := common.AccessModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	if osMode&os.O_RDWR != 0 {
		// open man says "The result is undefined if this flag is applied to a FIFO."
		// so, we don't allow it and return an error
		return nil, fmt.Errorf("O_READWRITE flag cannot be used for FIFO")
	}
	path := fifoPath(name)
	if mode&(ipc.O_OPEN_OR_CREATE|ipc.O_CREATE_ONLY) != 0 {
		err = unix.Mkfifo(path, uint32(perm))
		if err != nil {
			if mode&ipc.O_OPEN_OR_CREATE != 0 && os.IsExist(err) {
				err = nil
			} else {
				return nil, err
			}
		}
	}
	if mode&ipc.O_NONBLOCK != 0 {
		osMode |= unix.O_NONBLOCK
	}
	file, err := os.OpenFile(path, osMode, perm)
	if err != nil {
		return nil, err
	}
	return &fifoImpl{file: file}, nil
}

func (f *fifoImpl) Read(b []byte) (n int, err error) {
	return f.file.Read(b)
}

func (f *fifoImpl) Write(b []byte) (n int, err error) {
	return f.file.Write(b)
}

// Close closes the object
func (f *fifoImpl) Close() error {
	return f.file.Close()
}

// Destroy permanently removes the FIFO, closing it first.
func (f *fifoImpl) Destroy() error {
	var err error
	if err = f.file.Close(); err == nil {
		return os.Remove(f.file.Name())
	}
	return err
}

// Destroy permanently removes the FIFO.
func Destroy(name string) error {
	err := os.Remove(fifoPath(name))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// fifoPath returns full path for the fifo
// if its name contains '/' ('/tmp/fifo', './fifo') - use it
// if only filename was passed, assume it is in /tmp
func fifoPath(name string) string {
	if strings.Contains(name, "/") {
		return name
	}
	return "/tmp/" + name
}
