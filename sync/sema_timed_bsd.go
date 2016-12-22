// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package sync

import (
	"runtime"
	"time"

	"golang.org/x/sys/unix"
)

// WaitTimeout is supported on linux only.
func (s *semaphore) WaitTimeout(timeout time.Duration) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	tid := unix.Gettid()
	go func() {

	}()
	b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
	err := semop(s.id, []sembuf{b})
	panic(err)
}
