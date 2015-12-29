// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"
)

type rwMutexImpl struct {
}

func newRwMutexImpl(name string, mode int, perm os.FileMode) (impl *rwMutexImpl, resultErr error) {
	panic("unimplemented")
}
