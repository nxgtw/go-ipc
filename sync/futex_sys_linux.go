// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"math"
	"os"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/allocator"

	"golang.org/x/sys/unix"
)

const (
	cFUTEX_WAIT        = 0
	cFUTEX_WAKE        = 1
	cFUTEX_REQUEUE     = 3
	cFUTEX_CMP_REQUEUE = 4
	cFUTEX_WAKE_OP     = 5

	FUTEX_PRIVATE_FLAG   = 128
	FUTEX_CLOCK_REALTIME = 256

	cFutexWakeAll = math.MaxInt32
)

func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) (int32, error) {
	r1, _, err := unix.Syscall6(unix.SYS_FUTEX,
		uintptr(addr),
		uintptr(op),
		uintptr(val),
		uintptr(ts),
		uintptr(addr2),
		uintptr(val3))
	allocator.Use(addr)
	allocator.Use(addr2)
	if err != 0 {
		return 0, os.NewSyscallError("FUTEX", err)
	}
	return int32(r1), nil
}
