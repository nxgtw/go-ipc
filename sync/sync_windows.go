// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex  = modkernel32.NewProc("CreateMutexW")
	procOpenMutex    = modkernel32.NewProc("OpenMutexW")
	procReleaseMutex = modkernel32.NewProc("ReleaseMutex")
)

func createMutex(name string) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	h, _, err := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(namep)))
	if h == 0 {
		// do not wrap this error into a os.SyscallError, as
		// it can be check later for ERROR_ALREADY_EXISTS
		return windows.InvalidHandle, err
	}
	return windows.Handle(h), nil
}

func openMutex(name string) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	h, _, err := procOpenMutex.Call(uintptr(windows.SYNCHRONIZE), 0, uintptr(unsafe.Pointer(namep)))
	if h == 0 {
		return windows.InvalidHandle, os.NewSyscallError("OpenMutex", err)
	}
	return windows.Handle(h), nil
}

func releaseMutex(handle windows.Handle) error {
	result, _, err := procReleaseMutex.Call(uintptr(handle))
	if result == 0 {
		return os.NewSyscallError("ReleaseMutex", err)
	}
	return nil
}
