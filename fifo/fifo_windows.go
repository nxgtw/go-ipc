// Copyright 2015 Aleksandr Demakin. All rights reserved.

package fifo

import (
	"fmt"
	"os"

	"bitbucket.org/avd/go-ipc"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/windows"
)

type fifo struct {
	handle windows.Handle
}

func newFifo(name string, mode int, perm os.FileMode) (*fifo, error) {
	if _, err := common.CreateModeToOsMode(mode); err != nil {
		return nil, err
	}
	_, err := common.AccessModeToOsMode(mode)
	if err != nil {
		return nil, err
	}
	if mode&ipc.O_READWRITE != 0 {
		// open man says "The result is undefined if this flag is applied to a FIFO."
		// so, we don't allow it and return an error
		return nil, fmt.Errorf("O_READWRITE flag cannot be used for FIFO")
	}
	path := fifoPath(name)
	if mode&(ipc.O_CREATE_ONLY|ipc.O_OPEN_OR_CREATE) != 0 {
		var pipeMode uint32 = cPIPE_TYPE_MESSAGE | cPIPE_READMODE_MESSAGE
		if mode&ipc.O_NONBLOCK != 0 {
			pipeMode |= cPIPE_NOWAIT
		} else {
			pipeMode |= cPIPE_WAIT
		}
		handle, err := createNamedPipe(
			path,
			cPIPE_ACCESS_DUPLEX,
			pipeMode,
			cPIPE_UNLIMITED_INSTANCES,
			cFifoBufferSize,
			cFifoBufferSize,
			0,
			nil)
		_, _ = handle, err
	}
	return &fifo{}, nil
}

func fifoPath(name string) string {
	const template = `\\.\pipe\`
	return template + name
}
