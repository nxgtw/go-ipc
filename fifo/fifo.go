// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"io"
	"os"

	"bitbucket.org/avd/go-ipc"
)

// Fifo represents a First-In-First-Out object
type Fifo interface {
	io.ReadWriter
	io.Closer
	ipc.Destroyer
}

// New creates or opens a new FIFO object
//	name - object name.
//	mode - access mode. combination of the following flags:
//		O_OPEN_ONLY or O_CREATE_ONLY or O_OPEN_OR_CREATE
//		O_READ_ONLY or O_WRITE_ONLY
//		O_NONBLOCK.
//	perm - file permissions.
func New(name string, mode int, perm os.FileMode) (Fifo, error) {
	return newFifo(name, mode, perm)
}

// Destroy permanently removes the FIFO.
func Destroy(name string) error {
	return destroyFifo(name)
}
