// Copyright 2016 Aleksandr Demakin. All rights reserved.

// Package ipc provides primitives for inter-process communication.
// Currently it implements the following mechanisms:
//	fifo (unix, windows)
//	memory mapped files (unix, windows)
//	message queues (unix)
//	shared memory (unix, windows)
//	mutexes (unix, windows)
package ipc
