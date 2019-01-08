// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"fmt"
	"os"
	"time"

	"github.com/nxgtw/go-ipc/internal/common"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// NamedPipe is a first-in-first-out ipc mechanism.
type NamedPipe struct {
	pipeHandle windows.Handle
	name       string
}

// NewNamedPipe creates a new windows named pipe.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
func NewNamedPipe(name string, flag int, perm os.FileMode) (*NamedPipe, error) {
	if flag&os.O_RDWR != 0 {
		// open man says "The result is undefined if this flag is applied to a FIFO."
		// so, we don't allow it and return an error
		return nil, fmt.Errorf("O_RDWR flag cannot be used for FIFO")
	}
	path := namedPipePath(name)
	var pipeHandle windows.Handle
	var err error
	if flag&os.O_WRONLY != 0 {
		if pipeHandle, err = createFifoClient(path, flag); err != nil {
			return nil, err
		}
	} else {
		if pipeHandle, err = createFifoServer(path, flag); err != nil {
			return nil, err
		}
	}
	return &NamedPipe{pipeHandle: pipeHandle, name: name}, nil
}

// Read reads from the given FIFO. it must be opened for reading.
func (f *NamedPipe) Read(b []byte) (n int, err error) {
	var done uint32
	err = windows.ReadFile(f.pipeHandle, b, &done, nil)
	return int(done), err
}

// Write writes to the given FIFO. it must be opened for writing.
func (f *NamedPipe) Write(b []byte) (n int, err error) {
	var done uint32
	err = windows.WriteFile(f.pipeHandle, b, &done, nil)
	return int(done), err
}

// Close closes the object.
func (f *NamedPipe) Close() error {
	err := windows.CloseHandle(f.pipeHandle)
	f.pipeHandle = windows.InvalidHandle
	return err
}

// Destroy does the same as Close does.
// It is impossible to destroy named pipe explicitly,
// it will be destroyed by the OS when all its handles are closed.
func (f *NamedPipe) Destroy() error {
	return f.Close()
}

// DestroyNamedPipe is a no-op on windows.
func DestroyNamedPipe(name string) error {
	return nil
}

func namedPipePath(name string) string {
	const prefix = `\\.\pipe\`
	return prefix + name
}

func createFifoClient(path string, flag int) (windows.Handle, error) {
	namep, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windows.InvalidHandle, errors.Wrap(err, "invalid filename")
	}
	var fileHandle windows.Handle
	// unlike unix, we can't wait for a server to create a fifo,
	// if it doesn't exist. so, we are looping and waiting with a delay.
	const delay = time.Millisecond * 100
	for {
		fileHandle, err = windows.CreateFile(
			namep,
			windows.GENERIC_WRITE,
			0,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_ATTRIBUTE_NORMAL,
			0)
		if fileHandle != windows.InvalidHandle {
			break
		}
		if flag&O_NONBLOCK != 0 {
			return windows.InvalidHandle, errors.Wrap(err, "create file failed")
		}
		if os.IsNotExist(err) {
			time.Sleep(delay)
			continue
		}
		if !common.SyscallErrHasCode(os.NewSyscallError("CreateFile", err), cERROR_PIPE_BUSY) {
			return windows.InvalidHandle, errors.Wrap(err, "create file failed")
		}
		if flag&O_NONBLOCK != 0 {
			break
		}
		if ok, err := waitNamedPipe(path, cNMPWAIT_WAIT_FOREVER); !ok {
			return windows.InvalidHandle, errors.Wrap(err, "waitNamedPipe failed")
		}
	}

	newMode := uint32(cPIPE_READMODE_MESSAGE)
	if ok, err := setNamedPipeHandleState(fileHandle, &newMode, nil, nil); !ok {
		windows.CloseHandle(fileHandle)
		return windows.InvalidHandle, errors.Wrap(err, "setNamedPipeHandleState failed")
	}
	return fileHandle, nil
}

func createFifoServer(path string, flag int) (windows.Handle, error) {
	var pipeHandle = windows.InvalidHandle
	namep, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windows.InvalidHandle, errors.Wrap(err, "invalid filename")
	}
	creator := func(create bool) error {
		var err error
		if create {
			pipeHandle, err = makeNamedPipe(path, flag)
		} else {
			pipeHandle, err = windows.CreateFile(
				namep,
				windows.GENERIC_READ,
				0,
				nil,
				windows.OPEN_EXISTING,
				windows.FILE_ATTRIBUTE_NORMAL,
				0)
		}
		return err
	}
	for {
		_, err := common.OpenOrCreate(creator, flag)
		if pipeHandle == windows.InvalidHandle {
			return windows.InvalidHandle, errors.Wrap(err, "open/create file failed")
		}
		connected := true
		if flag&O_NONBLOCK == 0 {
			connected, err = connectNamedPipe(pipeHandle, nil)
			if !connected && common.SyscallErrHasCode(err, cERROR_PIPE_CONNECTED) {
				connected = true
			}
		}
		if connected {
			return pipeHandle, nil
		}
		if common.SyscallErrHasCode(err, cERROR_NO_DATA) {
			disconnectNamedPipe(pipeHandle)
		}
		windows.CloseHandle(pipeHandle)
	}
}

func makeNamedPipe(path string, flag int) (windows.Handle, error) {
	var pipeMode uint32 = cPIPE_TYPE_MESSAGE | cPIPE_READMODE_MESSAGE
	if flag&O_NONBLOCK != 0 {
		pipeMode |= cPIPE_NOWAIT
	} else {
		pipeMode |= cPIPE_WAIT
	}
	pipeHandle, err := createNamedPipe(
		path,
		cPIPE_ACCESS_DUPLEX,
		pipeMode,
		cPIPE_UNLIMITED_INSTANCES,
		cFifoBufferSize,
		cFifoBufferSize,
		0,
		nil)
	return pipeHandle, err
}
