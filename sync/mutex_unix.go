// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sync

import "os"

type mutexImpl struct {
}

func newMutexImpl(name string, mode int, perm os.FileMode) (*mutexImpl, error) {
	panic("unimplemented")
}

func (m *mutexImpl) Lock() {
	panic("unimplemented")
}

func (m *mutexImpl) Unlock() {
	panic("unimplemented")
}

func (m *mutexImpl) Close() error {
	panic("unimplemented")
}

func (m *mutexImpl) Destroy() error {
	panic("unimplemented")
}
