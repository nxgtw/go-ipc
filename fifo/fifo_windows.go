// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"os"

	"golang.org/x/sys/windows"
)

type fifoImpl struct {
	handle windows.Handle
}

func newFifoImpl(name string, mode int, perm os.FileMode) (*fifoImpl, error) {
	return &fifoImpl{}, nil
}
