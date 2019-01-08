// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux freebsd

package sync

import (
	"math"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/nxgtw/go-ipc/internal/common"
)

const (
	cFutexWakeAll = math.MaxInt32
)

type futex struct {
	ptr unsafe.Pointer
}

func (w *futex) addr() *int32 {
	return (*int32)(w.ptr)
}

func (w *futex) add(value int) {
	atomic.AddInt32(w.addr(), int32(value))
}

func (w *futex) wait(value int32, timeout time.Duration) error {
	err := FutexWait(w.ptr, value, timeout, 0)
	if err != nil && common.SyscallErrHasCode(err, unix.EWOULDBLOCK) {
		return nil
	}
	return err
}

func (w *futex) wake(count int32) (int, error) {
	return FutexWake(w.ptr, count, 0)
}

func (w *futex) wakeAll() (int, error) {
	return w.wake(cFutexWakeAll)
}
