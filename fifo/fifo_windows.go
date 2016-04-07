// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"os"

	"golang.org/x/sys/windows"
)

type fifo struct {
	handle windows.Handle
}

func newFifo(name string, mode int, perm os.FileMode) (*fifo, error) {
	//	path := fifoPath(name)
	return &fifo{}, nil
}

func fifoPath(name string) string {
	const template = `\\.\pipe\`
	return template + name
}
