// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

// Destroyer is an object which can be permanently removed.
type Destroyer interface {
	Destroy() error
}
