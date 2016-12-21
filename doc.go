// Copyright 2016 Aleksandr Demakin. All rights reserved.

// Package ipc provides primitives for inter-process communication.
// Works on Linux, OSX, FreeBSD, and Windows (x86 or x86-64).
// Supports following mechanisms:
//	fifo (all supported platforms)
//	memory mapped files (all supported platforms)
//	shared memory (all supported platforms)
//	system message queues (Linux, FreeBSD, OSX)
//	cross-platform priority message queue (all supported platforms)
//	mutexes, rw mutexes (all supported platforms)
//	semaphores (all supported platforms)
//	events (all supported platforms)
//	conditional variables (all supported platforms)
package ipc
