// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"golang.org/x/sys/windows"
)

const (
	cEVENT_MODIFY_STATE = 0x0002
)

var (
	modkernel32      = windows.NewLazyDLL("kernel32.dll")
	procCreateMutex  = modkernel32.NewProc("CreateMutexW")
	procOpenMutex    = modkernel32.NewProc("OpenMutexW")
	procReleaseMutex = modkernel32.NewProc("ReleaseMutex")
	procOpenEvent    = modkernel32.NewProc("OpenEventW")
	procCreateEvent  = modkernel32.NewProc("CreateEventW")
)

func createMutex(name string) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	h, _, err := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(namep)))
	allocator.Use(unsafe.Pointer(namep))
	if h == 0 {
		return 0, os.NewSyscallError("CreateMutex", err)
	}
	return windows.Handle(h), err
}

func openMutex(name string) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	h, _, err := procOpenMutex.Call(uintptr(windows.SYNCHRONIZE), 0, uintptr(unsafe.Pointer(namep)))
	allocator.Use(unsafe.Pointer(namep))
	if h == 0 {
		return 0, os.NewSyscallError("OpenMutex", err)
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

func openEvent(name string, desiredAccess uint32, inheritHandle uint32) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	h, _, err := procOpenEvent.Call(uintptr(desiredAccess), uintptr(inheritHandle), uintptr(unsafe.Pointer(namep)))
	allocator.Use(unsafe.Pointer(namep))
	if h == 0 {
		return 0, os.NewSyscallError("OpenEvent", err)
	}
	return windows.Handle(h), nil
}

func createEvent(name string, eventAttrs *windows.SecurityAttributes, manualReset uint32, initialState uint32) (handle windows.Handle, err error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	h, _, err := procCreateEvent.Call(
		uintptr(unsafe.Pointer(eventAttrs)),
		uintptr(manualReset),
		uintptr(initialState),
		uintptr(unsafe.Pointer(namep)))
	allocator.Use(unsafe.Pointer(eventAttrs))
	allocator.Use(unsafe.Pointer(namep))
	if h == 0 {
		err = os.NewSyscallError("CreateEvent", err)
	} else if err == syscall.Errno(0) {
		err = nil
	}
	return windows.Handle(h), err
}
