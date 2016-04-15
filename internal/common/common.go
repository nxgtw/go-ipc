// Copyright 2016 Aleksandr Demakin. All rights reserved.

package common

import (
	"syscall"

	"bitbucket.org/avd/go-ipc"

	"fmt"
	"os"
)

func OpenOrCreate(creator func(bool) error, mode int) (bool, error) {
	switch mode {
	case ipc.O_OPEN_ONLY:
		return false, creator(false)
	case ipc.O_CREATE_ONLY:
		err := creator(true)
		if err != nil {
			return false, err
		}
		return true, nil
	case ipc.O_OPEN_OR_CREATE:
		const attempts = 16
		var err error
		for attempt := 0; attempt < attempts; attempt++ {
			if err = creator(true); !os.IsExist(err) {
				return true, err
			}
			if err = creator(false); !os.IsNotExist(err) {
				return false, err
			}
		}
		return false, err
	default:
		return false, fmt.Errorf("unknown open mode")
	}
}

// AccessModeToOsMode converts library's access flags into
// os flags, which can be passed to system calls.
func AccessModeToOsMode(mode int) (osMode int, err error) {
	if mode&ipc.O_READ_ONLY != 0 {
		if mode&(ipc.O_WRITE_ONLY|ipc.O_READWRITE) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return osMode | os.O_RDONLY, nil
	}
	if mode&ipc.O_WRITE_ONLY != 0 {
		if mode&ipc.O_READWRITE != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return osMode | os.O_WRONLY, nil
	}
	if mode&ipc.O_READWRITE != 0 {
		return osMode | os.O_RDWR, nil
	}
	return 0, fmt.Errorf("no access mode flags")
}

// CreateModeToOsMode converts library's create flags into
// os flags, which can be passed to system calls.
func CreateModeToOsMode(mode int) (int, error) {
	if mode&ipc.O_OPEN_OR_CREATE != 0 {
		if mode&(ipc.O_CREATE_ONLY|ipc.O_OPEN_ONLY) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE, nil
	}
	if mode&ipc.O_CREATE_ONLY != 0 {
		if mode&ipc.O_OPEN_ONLY != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_EXCL, nil
	}
	if mode&ipc.O_OPEN_ONLY != 0 {
		return 0, nil
	}
	return 0, fmt.Errorf("no create mode flags")
}

func OpenModeToOsMode(mode int) (int, error) {
	var err error
	var createMode, accessMode int
	if createMode, err = CreateModeToOsMode(mode); err != nil {
		return 0, err
	}
	if accessMode, err = AccessModeToOsMode(mode); err != nil {
		return 0, err
	}
	return createMode | accessMode, nil
}

func SyscallErrHasCode(err error, code syscall.Errno) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno == code
		}
	}
	return false
}
