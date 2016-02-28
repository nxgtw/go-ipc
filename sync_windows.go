// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import "syscall"

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex  = modkernel32.NewProc("CreateMutexW")
	procOpenMutex    = modkernel32.NewProc("OpenMutexW")
	procReleaseMutex = modkernel32.NewProc("ReleaseMutex")
)
