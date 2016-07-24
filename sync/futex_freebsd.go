// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"math"
	"os"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"
	"bitbucket.org/avd/go-ipc/internal/common"

	"golang.org/x/sys/unix"
)

const (
	cUMTX_OP_WAIT_UINT         = 0xb
	cUMTX_OP_WAIT_UINT_PRIVATE = 0xf
	cUMTX_OP_WAKE              = 0x3
	cUMTX_OP_WAKE_PRIVATE      = 0x10
	cUMTX_ABSTIME              = 0x01

	cFutexWakeAll = math.MaxInt32
)

// FutexWait checks if the the value equals futex's value.
// If it doesn't, Wait returns EWOULDBLOCK.
// Otherwise, it waits for the Wake call on the futex for not longer, than timeout.
func FutexWait(addr unsafe.Pointer, value uint32, timeout time.Duration, flags int32) error {
	var ptr unsafe.Pointer
	if flags&cUMTX_ABSTIME != 0 {
		ptr = unsafe.Pointer(common.AbsTimeoutToTimeSpec(timeout))
	} else {
		ptr = unsafe.Pointer(common.TimeoutToTimeSpec(timeout))
	}
	fun := func() error {
		_, err := sys_umtx_op(addr, cUMTX_OP_WAIT_UINT|flags, value, nil, ptr)
		return err
	}
	return common.UninterruptedSyscall(fun)
}

// FutexWake wakes count threads waiting on the futex.
// Returns the number of woken threads.
func FutexWake(addr unsafe.Pointer, count uint32, flags int32) (int, error) {
	var woken int32
	fun := func() error {
		var err error
		woken, err = sys_umtx_op(addr, cUMTX_OP_WAKE|flags, count, nil, nil)
		return err
	}
	err := common.UninterruptedSyscall(fun)
	if err == nil {
		return int(woken), nil
	}
	return 0, err
}

func sys_umtx_op(addr unsafe.Pointer, mode int32, val uint32, ptr2, ts unsafe.Pointer) (int32, error) {
	r1, _, err := unix.Syscall6(unix.SYS__UMTX_OP,
		uintptr(addr),
		uintptr(mode),
		uintptr(val),
		uintptr(ptr2),
		uintptr(ts),
		0)
	allocator.Use(ptr2)
	allocator.Use(ts)
	if err != 0 {
		return 0, os.NewSyscallError("SYS__UMTX_OP", err)
	}
	return int32(r1), nil
}
