// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os"
	"unsafe"
)

func createModeToOsMode(mode int) (int, error) {
	if mode&O_OPEN_OR_CREATE != 0 {
		if mode&(O_CREATE_ONLY|O_OPEN_ONLY) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_TRUNC | os.O_RDWR, nil
	}
	if mode&O_CREATE_ONLY != 0 {
		if mode&O_OPEN_ONLY != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return os.O_CREATE | os.O_EXCL | os.O_RDWR, nil
	}
	if mode&O_OPEN_ONLY != 0 {
		return 0, nil
	}
	return 0, fmt.Errorf("no create mode flags")
}

func accessModeToOsMode(mode int) (umode int, err error) {
	if mode&O_READ_ONLY != 0 {
		if mode&(O_WRITE_ONLY|O_READWRITE) != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return umode | os.O_RDONLY, nil
	}
	if mode&O_WRITE_ONLY != 0 {
		if mode&O_READWRITE != 0 {
			return 0, fmt.Errorf("incompatible open flags")
		}
		return umode | os.O_WRONLY, nil
	}
	if mode&O_READWRITE != 0 {
		return umode | os.O_RDWR, nil
	}
	return 0, fmt.Errorf("no access mode flags")
}

func modeToOsMode(mode int) (int, error) {
	if createMode, err := createModeToOsMode(mode); err == nil {
		if accessMode, err := accessModeToOsMode(mode); err == nil {
			return createMode | accessMode, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}

// from syscall package
//go:noescape
func use(p unsafe.Pointer)
