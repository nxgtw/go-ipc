// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build linux windows

package sync

import (
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

type semaWaiter struct {
	s *semaphore
}

func newSemaWaiter(s Semaphore) *semaWaiter {
	return &semaWaiter{s: s.(*semaphore)}
}

func (sw *semaWaiter) wake(count uint32) (int, error) {
	sw.s.Signal(int(count))
	return int(count), nil
}

func (sw *semaWaiter) wait(unused uint32, timeout time.Duration) error {
	if !sw.s.WaitTimeout(timeout) {
		return common.NewTimeoutError("SEMWAWIT")
	}
	return nil
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
// This call is supported on linux and windows only.
func (m *SemaMutex) LockTimeout(timeout time.Duration) bool {
	return m.lwm.lockTimeout(timeout)
}
