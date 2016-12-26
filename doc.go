// Copyright 2016 Aleksandr Demakin. All rights reserved.

// Package ipc provides primitives for inter-process communication.
// Works on Linux, OSX, FreeBSD, and Windows (x86 or x86-64).
// Supports following mechanisms:
//	fifo (unix and windows pipes)
//	memory mapped files
//	shared memory
//	system message queues (Linux, FreeBSD, OSX)
//	cross-platform priority message queue
//	mutexes, rw mutexes
//	semaphores
//	events
//	conditional variables (all supported platforms)
package ipc
