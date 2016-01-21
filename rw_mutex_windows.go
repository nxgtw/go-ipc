// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"

	"golang.org/x/sys/windows"
)

type rwMutexImpl struct {
	handle windows.Handle
}

func newRwMutexImpl(name string, mode int, perm os.FileMode) (impl *rwMutexImpl, resultErr error) {
	if mode == O_OPEN_ONLY {

	}
	panic("unimplemented")
}
