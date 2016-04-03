// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

// Destroyer is an object which can be permanently removed
type Destroyer interface {
	Destroy() error
}

// Blocker is an object, whose operations can be blockable or not
type Blocker interface {
	SetBlocking(bool) error
}

// Buffered is an interface for objects with a capacity for storing other objects
type Buffered interface {
	Cap() (int, error)
}
