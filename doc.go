// Copyright 2016 Aleksandr Demakin. All rights reserved.

// Package ipc provides primitives for inter-process communication.
// Works on Linux, OSX, FreeBSD, and Windows.
// Supports following mechanisms:
//	fifo (all supported platforms)
//	memory mapped files (all supported platforms)
//	shared memory (all supported platforms)
//	system message queues (Linux, FreeBSD, OSX)
//	cross-platform priority message queue (all supported platforms)
//	locking primitives (all supported platforms)
//	conditional variables (Linux, FreeBSD, Windows, OSX)
package ipc
