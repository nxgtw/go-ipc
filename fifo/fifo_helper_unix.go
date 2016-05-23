// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd linux

package fifo

import "os"

func newFifo(name string, flag int, perm os.FileMode) (Fifo, error) {
	return NewUnixFifo(name, flag, perm)
}

func destroyFifo(name string) error {
	return DestroyUnixFIFO(name)
}
