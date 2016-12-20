// Copyright 2016 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"os"
	"syscall"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

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
	cERROR_PIPE_CONNECTED     = 535
	cERROR_PIPE_BUSY          = 231
	cERROR_NO_DATA            = 232
	cNMPWAIT_USE_DEFAULT_WAIT = 0x00000000
	cNMPWAIT_WAIT_FOREVER     = 0xffffffff
)

var (
	modkernel32                 = windows.NewLazyDLL("kernel32.dll")
	procCreateNamedPipe         = modkernel32.NewProc("CreateNamedPipeW")
	procConnectNamedPipe        = modkernel32.NewProc("ConnectNamedPipe")
	procWaitNamedPipe           = modkernel32.NewProc("WaitNamedPipeW")
	procSetNamedPipeHandleState = modkernel32.NewProc("SetNamedPipeHandleState")
	procDisconnectNamedPipe     = modkernel32.NewProc("DisconnectNamedPipe")
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
	allocator.Use(unsafe.Pointer(namep))
	allocator.Use(unsafe.Pointer(attrs))
	handle := windows.Handle(h)
	if handle == windows.InvalidHandle {
		return handle, os.NewSyscallError("CreateNamedPipe", err)
	}
	return handle, nil
}

func connectNamedPipe(handle windows.Handle, over *windows.Overlapped) (bool, error) {
	overPtr := unsafe.Pointer(over)
	r1, _, err := procConnectNamedPipe.Call(uintptr(handle), uintptr(overPtr))
	allocator.Use(overPtr)
	if err != syscall.Errno(0) {
		return false, os.NewSyscallError("ConnectNamedPipe", err)
	}
	return r1 != 0, nil
}

func disconnectNamedPipe(handle windows.Handle) (bool, error) {
	r1, _, err := procDisconnectNamedPipe.Call(uintptr(handle))
	if err != syscall.Errno(0) {
		return false, os.NewSyscallError("DisconnectNamedPipe", err)
	}
	return r1 != 0, nil
}

func waitNamedPipe(name string, timeout uint32) (bool, error) {
	namep, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return false, err
	}
	uNamep := unsafe.Pointer(namep)
	r1, _, err := procWaitNamedPipe.Call(uintptr(uNamep), uintptr(timeout))
	allocator.Use(uNamep)
	if err != syscall.Errno(0) {
		return false, os.NewSyscallError("WaitNamedPipe", err)
	}
	return r1 != 0, nil
}

func setNamedPipeHandleState(h windows.Handle, mode, maxCollectionCount, collectDataTimeout *uint32) (bool, error) {
	pMode := unsafe.Pointer(mode)
	pMaxCollectionCount, pCollectDataTimeout := unsafe.Pointer(maxCollectionCount), unsafe.Pointer(collectDataTimeout)
	r1, _, err := procSetNamedPipeHandleState.Call(
		uintptr(h),
		uintptr(pMode),
		uintptr(pMaxCollectionCount),
		uintptr(pCollectDataTimeout))
	allocator.Use(pMode)
	allocator.Use(pMaxCollectionCount)
	allocator.Use(pCollectDataTimeout)
	if err != syscall.Errno(0) {
		return false, os.NewSyscallError("SetNamedPipeHandleState", err)
	}
	return r1 != 0, nil
}
