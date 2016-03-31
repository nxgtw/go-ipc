// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/windows"
)

// SYSTEM_INFO is used for GetSystemInfo WinApi call
// see https://msdn.microsoft.com/en-us/library/windows/desktop/ms724958(v=vs.85).aspx
type SYSTEM_INFO struct {
	// This is the first member of the union
	OemID uint32
	// These are the second member of the union
	//      ProcessorArchitecture uint16;
	//      Reserved uint16;
	PageSize                  uint32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       *uint32
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
}

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemInfo   = modkernel32.NewProc("GetSystemInfo")
	procCreateNamedPipe = modkernel32.NewProc("CreateNamedPipeW")
)

func getAllocGranularity() int {
	var si SYSTEM_INFO
	ptr := unsafe.Pointer(&si)
	// this cannot fail
	procGetSystemInfo.Call(uintptr(ptr))
	allocator.Use(ptr)
	return int(si.AllocationGranularity)
}

func createNamedPipe(name string, openMode, pipeMode, maxInstances, outBufferSize, inBufferSize, defaultTimeout uint32,
	attrs *windows.SecurityAttributes) (windows.Handle, error) {
	_, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	//	procCreateNamedPipe.Call(namep)
	return windows.InvalidHandle, nil
}
