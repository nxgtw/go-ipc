// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sys

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/nxgtw/go-ipc/internal/allocator"

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
	modkernel32           = windows.NewLazyDLL("kernel32.dll")
	procGetSystemInfo     = modkernel32.NewProc("GetSystemInfo")
	procOpenFileMapping   = modkernel32.NewProc("OpenFileMappingW")
	procCreateFileMapping = modkernel32.NewProc("CreateFileMappingW")
)

// GetAllocGranularity returns system allocation granularity.
func GetAllocGranularity() int {
	var si systemInfo
	// this cannot fail
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	return int(si.AllocationGranularity)
}

// OpenFileMapping is a wraper for windows syscall.
func OpenFileMapping(access uint32, inheritHandle uint32, name string) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	nameu := unsafe.Pointer(namep)
	r1, _, err := procOpenFileMapping.Call(uintptr(access), uintptr(inheritHandle), uintptr(nameu))
	allocator.Use(nameu)
	if r1 == 0 {
		if err == windows.ERROR_FILE_NOT_FOUND {
			return 0, &os.PathError{Path: name, Op: "CreateFileMapping", Err: err}
		}
		return 0, os.NewSyscallError("OpenFileMapping", err)
	}
	if err == syscall.Errno(0) {
		err = nil
	}
	return windows.Handle(r1), err
}

// CreateFileMapping is a wraper for windwos syscall.
// We cannot use a call from golang.org/x/sys/windows, because it returns nil error, if the syscall returned a valid handle.
// However, CreateFileMapping may return a valid handle along with ERROR_ALREADY_EXISTS, and in this case
// we cannot find out, if the file existed before.
func CreateFileMapping(fhandle windows.Handle, sa *windows.SecurityAttributes, prot uint32, maxSizeHigh uint32, maxSizeLow uint32, name string) (handle windows.Handle, err error) {
	var namep *uint16
	if len(name) > 0 {
		namep, err = windows.UTF16PtrFromString(name)
		if err != nil {
			return 0, err
		}
	}
	nameu := unsafe.Pointer(namep)
	sau := unsafe.Pointer(sa)
	r1, _, err := procCreateFileMapping.Call(uintptr(fhandle), uintptr(sau), uintptr(prot), uintptr(maxSizeHigh), uintptr(maxSizeLow), uintptr(nameu))
	allocator.Use(sau)
	allocator.Use(nameu)
	if r1 == 0 {
		if err == windows.ERROR_ALREADY_EXISTS {
			return 0, &os.PathError{Path: name, Op: "CreateFileMapping", Err: err}
		}
		return 0, os.NewSyscallError("CreateFileMapping", err)
	}
	if err == syscall.Errno(0) {
		err = nil
	}
	return windows.Handle(r1), err
}
