// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"os"
	"unsafe"
)

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

// from syscall package:
// use is a no-op, but the compiler cannot see that it is.
// Calling use(p) ensures that p is kept live until that point.
//go:noescape
func use(p unsafe.Pointer)
