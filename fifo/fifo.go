// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"io"
	"os"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	// O_NONBLOCK flag makes Fifo open operation nonblocking.
	O_NONBLOCK = common.O_NONBLOCK
)

// Fifo represents a First-In-First-Out object
type Fifo interface {
	io.ReadWriter
	io.Closer
	Destroy() error
}

// New creates or opens a new FIFO object
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package along with O_NONBLOCK flag.
//	perm - object's permission bits.
func New(name string, flag int, perm os.FileMode) (Fifo, error) {
	return newFifo(name, flag, perm)
}

// Destroy permanently removes the FIFO.
func Destroy(name string) error {
	return destroyFifo(name)
}
