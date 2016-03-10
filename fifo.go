// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import "os"

// Fifo represents a First-in-First Out object
type Fifo struct {
	*fifoImpl
}

// NewFifo creates or opens new FIFO object
// name - object name.
// mode - access mode. can be one of the following:
//  O_OPEN_ONLY or O_CREATE_ONLY or O_OPEN_OR_CREATE
//	O_READ_ONLY or	O_WRITE_ONLY
//  some platform-specific flags can be used as well
// perm - file permissions
func NewFifo(name string, mode int, perm os.FileMode) (*Fifo, error) {
	impl, err := newFifoImpl(name, mode, perm)
	if err != nil {
		return nil, err
	}
	return &Fifo{fifoImpl: impl}, err
}
