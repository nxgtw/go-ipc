// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

// common flags for opening/creation of objects
const (
	O_OPEN_OR_CREATE = 0x00000001
	O_CREATE_ONLY    = 0x00000002
	O_OPEN_ONLY      = 0x00000004
	O_READ_ONLY      = 0x00000008
	O_WRITE_ONLY     = 0x00000010
	O_READWRITE      = 0x00000020
	// O_NONBLOCK flag makes some ipc operations non-blocking
	O_NONBLOCK = 0x00000040
	// other values can be platform-specific, and/or operation-specific
)

// constants for memory regions
const (
	MEM_READ_ONLY     = 0x00000001
	MEM_READ_PRIVATE  = 0x00000002
	MEM_READWRITE     = 0x00000004
	MEM_COPY_ON_WRITE = 0x00000008
)
