// Copyright 2016 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	cPIPE_ACCESS_DUPLEX       = 0x00000003
	cPIPE_TYPE_MESSAGE        = 0x00000004
	cPIPE_READMODE_MESSAGE    = 0x00000002
	cPIPE_WAIT                = 0x00000000
	cPIPE_NOWAIT              = 0x00000001
	cPIPE_UNLIMITED_INSTANCES = 255
	cFifoBufferSize           = 512
)

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procCreateNamedPipe = modkernel32.NewProc("CreateNamedPipeW")
)

func createNamedPipe(name string, openMode, pipeMode, maxInstances, outBufferSize, inBufferSize, defaultTimeout uint32,
	attrs *windows.SecurityAttributes) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	h, _, err := procCreateNamedPipe.Call(
		uintptr(unsafe.Pointer(namep)),
		uintptr(openMode),
		uintptr(pipeMode),
		uintptr(maxInstances),
		uintptr(outBufferSize),
		uintptr(inBufferSize),
		uintptr(defaultTimeout),
		uintptr(unsafe.Pointer(attrs)))
	handle := windows.Handle(h)
	if handle == windows.InvalidHandle {
		return handle, os.NewSyscallError("CreateNamedPipe", err)
	}
	return handle, nil
}
