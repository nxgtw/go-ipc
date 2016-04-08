// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

// Destroyer is an object which can be permanently removed.
type Destroyer interface {
	Destroy() error
}

// Blocker is an object, which can work in blocking and non-blocking modes.
type Blocker interface {
	SetBlocking(bool) error
}

// Buffered is an object with internal buffer of the given capacity.
type Buffered interface {
	Cap() (int, error)
}
