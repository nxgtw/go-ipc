// Copyright 2016 Aleksandr Demakin. All rights reserved.

package common

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

const (
	// O_NONBLOCK flag tell some functions not to block.
	// Its value does not interfere with O_* constants from 'os' package.
	O_NONBLOCK = syscall.O_NONBLOCK
)

// Destroyer is an object which can be permanently removed.
type Destroyer interface {
	Destroy() error
}

// FlagsForOpen extracts os.O_CREATE and os.O_EXCL flag values.
func FlagsForOpen(flag int) int {
	return flag & (os.O_CREATE | os.O_EXCL)
}

// FlagsForAccess extracts os.O_RDONLY, os.O_WRONLY, os.O_RDWR, and O_NONBLOCK flag values.
func FlagsForAccess(flag int) int {
	return flag & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR | O_NONBLOCK)
}

// OpenOrCreate performs open/create file operation according to the given mode.
// It allows to find out if the object was opened or created.
//	creator is the function which performs actual operation:
//		if is called with 'true', if it must create an object, and with false otherwise.
//		it must return an 'not exists error' if the param is false, and the object does not exist.
//		it must return an 'already exists error' if the param is true, and the object already exists.
//	flag is the combination of open flags from os package.
//		If flag == os.O_CREATE, OpenOrCreate makes several attempts to open or create an object,
//		and analyzes the return error. It tries to open the object first.
func OpenOrCreate(creator func(bool) error, flag int) (bool, error) {
	flag = FlagsForOpen(flag)
	switch flag {
	case 0:
		return false, creator(false)
	case os.O_CREATE | os.O_EXCL:
		err := creator(true)
		if err != nil {
			return false, err
		}
		return true, nil
	case os.O_CREATE:
		const attempts = 16
		var err error
		for attempt := 0; attempt < attempts; attempt++ {
			if err = creator(false); !os.IsNotExist(err) {
				return false, err
			}
			if err = creator(true); !os.IsExist(err) {
				return true, err
			}
		}
		return false, err
	default:
		return false, fmt.Errorf("unknown open mode")
	}
}

// SyscallErrHasCode returns true, if given error is a syscall error with given code.
func SyscallErrHasCode(err error, code syscall.Errno) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno == code
		}
	}
	return false
}

// CallTimeout calls f in a loop allowing it to run for at least 'timeout'.
// It calls f, measuring its runtime:
//	if f returned false, or its cumulative runtime exceeded timeout, CallTimeout returns.
//	otherwise CallTimeout subtracts runtime from timeout anf calls 'f' with the updated value.
func CallTimeout(f func(time.Duration) bool, timeout time.Duration) {
	for {
		opStart := time.Now()
		if !f(timeout) {
			return
		}
		if timeout >= 0 {
			elapsed := time.Since(opStart)
			if timeout > elapsed {
				timeout = timeout - elapsed
			} else {
				return
			}
		}
	}
}
