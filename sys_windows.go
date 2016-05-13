// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"syscall"
	"unsafe"
)

// systemInfo is used for GetSystemInfo WinApi call
// see https://msdn.microsoft.com/en-us/library/windows/desktop/ms724958(v=vs.85).aspx
type systemInfo struct {
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
	modkernel32       = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemInfo = modkernel32.NewProc("GetSystemInfo")
)

func getAllocGranularity() int {
	var si systemInfo
	// this cannot fail
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	return int(si.AllocationGranularity)
}
