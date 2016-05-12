// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import "os"

// Fifo represents a First-In-First-Out object
type Fifo struct {
	*fifo
}

// New creates or opens a new FIFO object
//	name - object name.
//	mode - access mode. combination of the following flags:
//		O_OPEN_ONLY or O_CREATE_ONLY or O_OPEN_OR_CREATE
//		O_READ_ONLY or O_WRITE_ONLY
//		O_NONBLOCK.
//	perm - file permissions.
func New(name string, mode int, perm os.FileMode) (*Fifo, error) {
	impl, err := newFifo(name, mode, perm)
	if err != nil {
		return nil, err
	}
	return &Fifo{fifo: impl}, err
}

// Read reads from the given FIFO. The FIFO must be opened for reading.
func (f *Fifo) Read(b []byte) (n int, err error) {
	return f.fifo.Read(b)
}

// Write writes to the given FIFO. The FIFO must be opened for writing.
func (f *Fifo) Write(b []byte) (n int, err error) {
	return f.fifo.Write(b)
}

// Close closes the object.
func (f *Fifo) Close() error {
	return f.fifo.Close()
}

// Destroy permanently removes the FIFO, closing it first.
func (f *Fifo) Destroy() error {
	return f.fifo.Destroy()
}
