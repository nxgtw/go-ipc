// Copyright 2016 Aleksandr Demakin. All rights reserved.

package common

import (
	"os"
	"syscall"
)

const (
	ERROR_TIMEOUT = 1460
)

// IsTimeoutErr returns true, if the given error is a temporary syscall error.
func IsTimeoutErr(err error) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno == ERROR_TIMEOUT
		}
	}
	return false
}
