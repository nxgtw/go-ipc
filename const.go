// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

// common flags for opening/creation of objects
const (
	// flags below are common for all the platforms and open operations
	O_OPEN_OR_CREATE = 0x00000001
	O_CREATE_ONLY    = 0x00000002
	O_OPEN_ONLY      = 0x00000004
	O_READ_ONLY      = 0x00000008
	O_WRITE_ONLY     = 0x00000010
	O_READWRITE      = 0x00000020
	// other values can be platform-specific, and/or operation-specific
)

// constants for shared memory regions
const (
	SHM_READ_ONLY     = 0x00000001
	SHM_READ_PRIVATE  = 0x00000002
	SHM_READWRITE     = 0x00000004
	SHM_COPY_ON_WRITE = 0x00000008
)
