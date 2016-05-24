// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package fifo

import (
	"os"

	"bitbucket.org/avd/go-ipc/internal/common"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// UnixFifo is a first-in-first-out unix ipc mechanism.
type UnixFifo struct {
	file *os.File
}

// NewUnixFifo creates a new unix FIFO.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm = object permissions.
func NewUnixFifo(name string, flag int, perm os.FileMode) (*UnixFifo, error) {
	if flag&os.O_RDWR != 0 {
		// open man says "The result is undefined if this flag is applied to a FIFO."
		// so, we don't allow it and return an error
		return nil, errors.Errorf("O_RDWR flag cannot be used for FIFO")
	}
	path := fifoPath(name)
	var file *os.File
	creator := func(create bool) error {
		var err error
		if create {
			err = unix.Mkfifo(path, uint32(perm))
		}
		if err == nil {
			file, err = os.OpenFile(path, common.FlagsForAccess(flag), perm)
		}
		return err
	}
	if _, err := common.OpenOrCreate(creator, flag); err != nil {
		return nil, errors.Wrap(err, "open/create fifo failed")
	}
	return &UnixFifo{file: file}, nil
}

// Read reads from the given FIFO. it must be opened for reading.
func (f *UnixFifo) Read(b []byte) (n int, err error) {
	return f.file.Read(b)
}

// Write writes to the given FIFO. it must be opened for writing.
func (f *UnixFifo) Write(b []byte) (n int, err error) {
	return f.file.Write(b)
}

// Close closes the object.
func (f *UnixFifo) Close() error {
	return f.file.Close()
}

// Destroy permanently removes the FIFO, closing it first.
func (f *UnixFifo) Destroy() error {
	var err error
	err = f.file.Close()
	if err != nil {
		return errors.Wrap(err, "close failed")
	}
	if err = os.Remove(f.file.Name()); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "remove failed")
		}
	}
	return nil
}

// DestroyUnixFIFO permanently removes the FIFO.
func DestroyUnixFIFO(name string) error {
	err := os.Remove(fifoPath(name))
	if os.IsNotExist(err) {
		return nil
	}
	return errors.Wrap(err, "remove failed")
}

func fifoPath(name string) string {
	return "/tmp/" + name
}
