// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"math"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"bitbucket.org/avd/go-ipc/internal/common"
)

const (
	futexSize     = int(unsafe.Sizeof(uint32(0)))
	cFutexWakeAll = math.MaxInt32
)

type futex struct {
	ptr unsafe.Pointer
}

func (w *futex) addr() *uint32 {
	return (*uint32)(w.ptr)
}

func (w *futex) add(value int) {
	atomic.AddUint32(w.addr(), uint32(value))
}

func (w *futex) wait(value uint32, timeout time.Duration) error {
	err := FutexWait(w.ptr, value, timeout, 0)
	if err != nil && common.SyscallErrHasCode(err, syscall.EWOULDBLOCK) {
		return nil
	}
	return err
}

func (w *futex) wake(count uint32) (int, error) {
	return FutexWake(w.ptr, count, 0)
}

func (w *futex) wakeAll() (int, error) {
	return w.wake(cFutexWakeAll)
}
