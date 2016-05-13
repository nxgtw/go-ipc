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
	IpcCreate = 00001000 /* create if key is nonexistent */
	IpcExcl   = 00002000 /* fail if key exists */
	IpcNoWait = 00004000 /* return error on wait */

	IpcRmid = 0 /* remove resource */
	IpcSet  = 1 /* set ipc_perm options */
	IpcStat = 2 /* get ipc_perm options */
	IpcInfo = 3 /* see ipcs */
)

type Key uint64

func KeyForName(name string) (Key, error) {
	name = TmpFilename(name)
	file, err := os.Create(name)
	if err != nil {
		return 0, errors.New("invalid name for key")
	}
	file.Close()
	k, err := Ftok(name)
	if err != nil {
		os.Remove(name)
		return 0, errors.New("invalid name for key")
	}
	return k, nil
}

func TmpFilename(name string) string {
	return os.TempDir() + "/" + name
}

func Ftok(name string) (Key, error) {
	var statfs unix.Stat_t
	if err := unix.Stat(name, &statfs); err != nil {
		return Key(0), err
	}
	// unconvert says there is 'redundant type conversion' to uint64,
	// however, this is not always true, as the types of statfs.Ino and statfs.Dev
	// may vary on different platforms
	return Key(uint64(statfs.Ino)&0xFFFF | ((uint64(statfs.Dev) & 0xFF) << 16)), nil
}

func AbsTimeoutToTimeSpec(timeout time.Duration) *unix.Timespec {
	if timeout >= 0 {
		ts := unix.NsecToTimespec(time.Now().Add(timeout).UnixNano())
		return &ts
	}
	return nil
}

func TimeoutToTimeSpec(timeout time.Duration) *unix.Timespec {
	if timeout >= 0 {
		ts := unix.NsecToTimespec(timeout.Nanoseconds())
		return &ts
	}
	return nil
}

func IsInterruptedSyscallErr(err error) bool {
	return SyscallErrHasCode(err, syscall.EINTR)
}

func IsTimeoutErr(err error) bool {
	if sysErr, ok := err.(*os.SyscallError); ok {
		if errno, ok := sysErr.Err.(syscall.Errno); ok {
			return errno.Timeout()
		}
	}
	return false
}

func SyscallNameFromErr(err error) string {
	if sysErr, ok := err.(*os.SyscallError); ok {
		return sysErr.Syscall
	}
	return ""
}

func UninterruptedSyscall(f func() error) error {
	for {
		err := f()
		if !IsInterruptedSyscallErr(err) {
			return err
		}
	}
}

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
