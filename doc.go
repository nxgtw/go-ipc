// Copyright 2016 Aleksandr Demakin. All rights reserved.

// Package ipc provides primitives for inter-process communication.
// Currently it implements the following mechanisms:
//	fifo (unix, windows)
//	memory mapped files (unix, windows)
//	system message queues (unix)
//	shared memory (unix, windows)
//	mutexes (unix, windows)
// The library is currently alpha, so it lacks docs and examples,
// I'll add them later. Also, its public API may change.
// You can find some usage examples in test files of the subpackages.
package ipc
