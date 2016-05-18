// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package fifo

import "os"

func newFifo(name string, mode int, perm os.FileMode) (Fifo, error) {
	return NewUnixFifo(name, mode, perm)
}

func destroyFifo(name string) error {
	return DestroyUnixFIFO(name)
}
