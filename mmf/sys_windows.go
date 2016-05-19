// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mmf

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/windows"
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
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemInfo   = modkernel32.NewProc("GetSystemInfo")
	procOpenFileMapping = modkernel32.NewProc("OpenFileMappingW")
)

func getAllocGranularity() int {
	var si systemInfo
	// this cannot fail
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	return int(si.AllocationGranularity)
}

func openFileMapping(access uint32, inheritHandle uint32, name string) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	nameu := unsafe.Pointer(namep)
	r1, _, err := procOpenFileMapping.Call(uintptr(access), uintptr(inheritHandle), uintptr(nameu))
	allocator.Use(nameu)
	if r1 == 0 {
		return 0, os.NewSyscallError("OpenFileMapping", err)
	}
	if err == syscall.Errno(0) {
		err = nil
	}
	return windows.Handle(r1), err
}
