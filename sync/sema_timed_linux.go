// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

// WaitTimeout decrements the value of semaphore variable by 1.
// If the value becomes negative, it waites for not longer than timeout.
// This call is supported on linux only.
func (s *Semaphore) WaitTimeout(timeout time.Duration) {
	err := common.UninterruptedSyscallTimeout(func(curTimeout time.Duration) error {
		b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
		return semtimedop(s.id, []sembuf{b}, common.TimeoutToTimeSpec(curTimeout))
	}, timeout)
	if err != nil {
		panic(err)
	}
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
// This call is supported on linux only.
func (m *SemaMutex) LockTimeout(timeout time.Duration) bool {
	return m.inplace.lockTimeout(timeout)
}

type semaTimedWaiter struct {
	s *Semaphore
}

func newSemaWaiter(s *Semaphore) *semaTimedWaiter {
	return &semaTimedWaiter{s: s}
}

func (sw *semaTimedWaiter) wake() {
	sw.s.Signal(1)
}

func (sw *semaTimedWaiter) wait(timeout time.Duration) {
	sw.s.WaitTimeout(timeout)
}
