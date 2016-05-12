// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	futexSize = 4
)

// Futex is a linux ipc mechanism, which can be used to implement different synchronization objects.
type Futex struct {
	uaddr unsafe.Pointer
}

func NewFutex(uaddr unsafe.Pointer) *Futex {
	return &Futex{uaddr: uaddr}
}

// Addr returns address of the futex's value.
func (f *Futex) Addr() *uint32 {
	return (*uint32)(f.uaddr)
}

// Wait checks if the the value equals futex's value.
// If it doesn't, Wait returns EWOULDBLOCK.
// Otherwise, it waits for the Wake call on the futex for not longer, than timeout.
func (f *Futex) Wait(value uint32, timeout time.Duration, flags int32) error {
	var ptr unsafe.Pointer
	if flags&FUTEX_CLOCK_REALTIME != 0 {
		ptr = unsafe.Pointer(common.AbsTimeoutToTimeSpec(timeout))
	} else {
		ptr = unsafe.Pointer(common.TimeoutToTimeSpec(timeout))
	}
	fun := func() error {
		_, err := futex(f.uaddr, cFUTEX_WAIT|flags, value, ptr, nil, 0)
		return err
	}
	return common.UninterruptedSyscall(fun)
}

// Wake wakes count threads waiting on the futex.
func (f *Futex) Wake(count uint32, flags int32) (int, error) {
	var woken int32
	fun := func() error {
		var err error
		woken, err = futex(f.uaddr, cFUTEX_WAKE|flags, count, nil, nil, 0)
		return err
	}
	err := common.UninterruptedSyscall(fun)
	if err == nil {
		return int(woken), nil
	}
	return 0, err
}
