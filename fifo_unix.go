// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package ipc

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type fifoImpl struct {
	file *os.File
}

// NewFifo creates or opens new FIFO object
// name - object name. if it does not contain '/', then '/tmp/' prefix will be added
// mode - access mode. can be one of the following:
//	O_READ_ONLY
//	O_WRITE_ONLY
//	O_FIFO_NONBLOCK can be used with O_READ_ONLY
// perm - file permissions
func newFifoImpl(name string, mode int, perm os.FileMode) (*fifoImpl, error) {
	path := fifoPath(name)
	if err := unix.Mkfifo(path, uint32(perm)); err != nil && !os.IsExist(err) {
		return nil, err
	}
	osMode, err := accessModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	if osMode&os.O_RDWR != 0 {
		// open man says "The result is undefined if this flag is applied to a FIFO."
		// so, we don't allow it and return an error
		return nil, fmt.Errorf("O_READWRITE flag cannot be used for FIFO")
	}
	if mode&O_NONBLOCK != 0 {
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

// Destroy permanently removes the object, closing it at first.
func (f *fifoImpl) Destroy() error {
	var err error
	if err = f.file.Close(); err == nil {
		return os.Remove(f.file.Name())
	}
	return err
}

// DestroyFifo permanently removes the object.
func DestroyFifo(name string) error {
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
