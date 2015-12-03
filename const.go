// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

// common flags for opening/creation of objects
const (
	O_OPEN_OR_CREATE = 1 << iota
	O_CREATE_ONLY
	O_OPEN_ONLY
	O_READ_ONLY
	O_WRITE_ONLY
	O_READWRITE
	O_NONBLOCK // for FIFO open only
)

// constants for shared memory regions
const (
	SHM_READ_ONLY = iota
	SHM_READ_PRIVATE
	SHM_READWRITE
	SHM_COPY_ON_WRITE
)
