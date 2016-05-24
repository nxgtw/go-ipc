// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package common

import (
	"errors"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// IpcCreate flag tells a function to create an object if key is nonexistent.
	IpcCreate = 00001000
	// IpcExcl flag tells a function to create an object if key is nonexistent and fail if key exists.
	IpcExcl = 00002000
	// IpcNoWait flag tell a function to return error on wait.
	IpcNoWait = 00004000

	// IpcRmid flag tells a function to remove resource.
	IpcRmid = 0
	// IpcSet flag tells a function to set ipc_perm options.
	IpcSet = 1
	// IpcStat flag tells a function to get ipc_perm options.
	IpcStat = 2
	// IpcInfo flag tells a function to retrieve information about an object.
	IpcInfo = 3
)

// Key is an unsigned integer value considered to be unique for a unique name.
type Key uint64

// KeyForName generates a key for given path.
func KeyForName(name string) (Key, error) {
	name = TmpFilename(name)
	file, err := os.Create(name)
	if err != nil {
		return 0, errors.New("invalid name for key")
	}
	file.Close()
	k, err := ftok(name)
	if err != nil {
		os.Remove(name)
		return 0, errors.New("invalid name for key")
	}
	return k, nil
}

// TmpFilename returns a full path for a temporary file with the given name.
func TmpFilename(name string) string {
	return os.TempDir() + "/" + name
}

// AbsTimeoutToTimeSpec converts given timeout value to absulute value of unix.Timespec.
func AbsTimeoutToTimeSpec(timeout time.Duration) *unix.Timespec {
	if timeout >= 0 {
		ts := unix.NsecToTimespec(time.Now().Add(timeout).UnixNano())
		return &ts
	}
	return nil
}

// TimeoutToTimeSpec converts given timeout value to relative value of unix.Timespec.
func TimeoutToTimeSpec(timeout time.Duration) *unix.Timespec {
	if timeout >= 0 {
		ts := unix.NsecToTimespec(timeout.Nanoseconds())
		return &ts
	}
	return nil
}

// IsInterruptedSyscallErr returns true, if the given error is a syscall.EINTR error.
func IsInterruptedSyscallErr(err error) bool {
	return SyscallErrHasCode(err, syscall.EINTR)
}

// IsTimeoutErr returns true, if the given error is a temporary syscall.
func IsTimeoutErr(err error) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno.Timeout()
		}
	}
	return false
}

// SyscallNameFromErr returns name of a syscall from a syscall errror.
func SyscallNameFromErr(err error) string {
	if sysErr, ok := err.(*os.SyscallError); ok {
		return sysErr.Syscall
	}
	return ""
}

// UninterruptedSyscall runs a function in a loop.
// If an error, returned by the function is a syscall.EINTR error,
// it runs the function again. Otherwise, it returns the error.
func UninterruptedSyscall(f func() error) error {
	for {
		err := f()
		if !IsInterruptedSyscallErr(err) {
			return err
		}
	}
}

// UninterruptedSyscallTimeout runs a function in a loop.
// It acts like UninterruptedSyscall, however, before every run it
// recalculates timeout value according to the passed time.
func UninterruptedSyscallTimeout(f func(time.Duration) error, timeout time.Duration) error {
	for {
		opStart := time.Now()
		err := f(timeout)
		if !IsInterruptedSyscallErr(err) {
			return err
		}
		if timeout >= 0 {
			// we were interrupted by a signal. recalculate timeout
			elapsed := time.Now().Sub(opStart)
			if timeout > elapsed {
				timeout = timeout - elapsed
			} else {
				return os.NewSyscallError(SyscallNameFromErr(err), syscall.EAGAIN)
			}
		}
	}
}

func ftok(name string) (Key, error) {
	var statfs unix.Stat_t
	if err := unix.Stat(name, &statfs); err != nil {
		return Key(0), err
	}
	// unconvert says there is 'redundant type conversion' to uint64,
	// however, this is not always true, as the types of statfs.Ino and statfs.Dev
	// may vary on different platforms
	return Key(uint64(statfs.Ino)&0xFFFF | ((uint64(statfs.Dev) & 0xFF) << 16)), nil
}
