// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"
	"golang.org/x/sys/windows"
)

const (
	cEVENT_MODIFY_STATE = 0x0002
)

var (
	modkernel32          = windows.NewLazyDLL("kernel32.dll")
	procCreateMutex      = modkernel32.NewProc("CreateMutexW")
	procOpenMutex        = modkernel32.NewProc("OpenMutexW")
	procReleaseMutex     = modkernel32.NewProc("ReleaseMutex")
	procOpenEvent        = modkernel32.NewProc("OpenEventW")
	procCreateEvent      = modkernel32.NewProc("CreateEventW")
	procCreateSemaphore  = modkernel32.NewProc("CreateSemaphoreW")
	procOpenSemaphore    = modkernel32.NewProc("OpenSemaphoreW")
	procReleaseSemaphore = modkernel32.NewProc("ReleaseSemaphore")
)

func sys_CreateMutex(name string) (windows.Handle, error) {
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

func sys_OpenMutex(name string) (windows.Handle, error) {
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

func sys_releaseMutex(handle windows.Handle) error {
	result, _, err := procReleaseMutex.Call(uintptr(handle))
	if result == 0 {
		return os.NewSyscallError("ReleaseMutex", err)
	}
	return nil
}

func sys_OpenEvent(name string, desiredAccess uint32, inheritHandle uint32) (windows.Handle, error) {
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

func sys_CreateEvent(name string, eventAttrs *windows.SecurityAttributes, manualReset uint32, initialState uint32) (handle windows.Handle, err error) {
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

func openOrCreateEvent(name string, flag int, initial int) (windows.Handle, error) {
	var handle windows.Handle
	creator := func(create bool) error {
		var err error
		if create {
			handle, err = sys_CreateEvent(name, nil, 0, uint32(initial))
			if os.IsExist(err) {
				windows.CloseHandle(handle)
				return err
			}
		} else {
			handle, err = sys_OpenEvent(name, windows.SYNCHRONIZE|cEVENT_MODIFY_STATE, uint32(0))
		}
		if handle != windows.Handle(0) {
			return nil
		}
		return err
	}
	_, err := common.OpenOrCreate(creator, flag)
	return handle, err
}

func sys_CreateSemaphore(name string, initial, maximum int, attrs *windows.SecurityAttributes) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	h, _, err := procCreateEvent.Call(
		uintptr(unsafe.Pointer(attrs)),
		uintptr(initial),
		uintptr(maximum),
		uintptr(unsafe.Pointer(namep)))
	allocator.Use(unsafe.Pointer(attrs))
	allocator.Use(unsafe.Pointer(namep))
	if h == 0 {
		err = os.NewSyscallError("CreateSemaphore", err)
	} else if err == syscall.Errno(0) {
		err = nil
	}
	return windows.Handle(h), err
}

func sys_ReleaseSemaphore(h windows.Handle, count int) (int, error) {
	var prev int32
	prevPtr := unsafe.Pointer(&prev)
	ok, _, err := procReleaseSemaphore.Call(
		uintptr(h),
		uintptr(count),
		uintptr(prevPtr),
	)
	allocator.Use(prevPtr)
	if ok == 0 {
		err = os.NewSyscallError("ReleaseSemaphore", err)
	} else {
		err = nil
	}
	return int(prev), err
}

func sys_OpenSemaphore(name string, desiredAccess uint32, inheritHandle uint32) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	h, _, err := procOpenSemaphore.Call(uintptr(desiredAccess), uintptr(inheritHandle), uintptr(unsafe.Pointer(namep)))
	allocator.Use(unsafe.Pointer(namep))
	if h == 0 {
		return 0, os.NewSyscallError("OpenSemaphore", err)
	}
	return windows.Handle(h), nil
}

func openOrCreateSemaphore(name string, flag int, initial, maximum int) (windows.Handle, error) {
	var handle windows.Handle
	creator := func(create bool) error {
		var err error
		if create {
			handle, err = sys_CreateSemaphore(name, initial, maximum, nil)
			if os.IsExist(err) {
				windows.CloseHandle(handle)
				return err
			}
		} else {
			handle, err = sys_OpenSemaphore(name, windows.SYNCHRONIZE|cEVENT_MODIFY_STATE, uint32(0))
		}
		if handle != windows.Handle(0) {
			return nil
		}
		return err
	}
	_, err := common.OpenOrCreate(creator, flag)
	return handle, err
}
