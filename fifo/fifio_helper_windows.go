// Copyright 2016 Aleksandr Demakin. All rights reserved.

package fifo

import "os"

func newFifo(name string, mode int, perm os.FileMode) (Fifo, error) {
	return NewNamedPipe(name, mode, perm)
}

func destroyFifo(name string) error {
	return DestroyNamedPipe(name)
}
