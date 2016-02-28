// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"os"

	"golang.org/x/sys/windows"
)

type mutexImpl struct {
	handle windows.Handle
}

func newMutexImpl(name string, mode int, perm os.FileMode) (impl *mutexImpl, resultErr error) {
	switch mode {
	case O_OPEN_ONLY:
	case O_CREATE_ONLY:
	case O_OPEN_OR_CREATE:
	}
	panic("unimplemented")
}
