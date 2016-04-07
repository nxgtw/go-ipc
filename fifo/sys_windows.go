// Copyright 2016 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procCreateNamedPipe = modkernel32.NewProc("CreateNamedPipeW")
)

func createNamedPipe(name string, openMode, pipeMode, maxInstances, outBufferSize, inBufferSize, defaultTimeout uint32,
	attrs *windows.SecurityAttributes) (windows.Handle, error) {
	_, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	//	procCreateNamedPipe.Call(namep)
	return windows.InvalidHandle, nil
}
