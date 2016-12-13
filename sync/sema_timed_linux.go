// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
	"golang.org/x/sys/unix"
)

// WaitTimeout is supported on linux only.
func (s *semaphore) WaitTimeout(timeout time.Duration) bool {
	err := common.UninterruptedSyscallTimeout(func(curTimeout time.Duration) error {
		b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
		return semtimedop(s.id, []sembuf{b}, common.TimeoutToTimeSpec(curTimeout))
	}, timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
// This call is supported on linux only.
func (m *SemaMutex) LockTimeout(timeout time.Duration) bool {
	return m.lwm.lockTimeout(timeout)
}

type semaTimedWaiter struct {
	s *semaphore
}

func newSemaWaiter(s Semaphore) *semaTimedWaiter {
	return &semaTimedWaiter{s: s.(*semaphore)}
}

func (sw *semaTimedWaiter) wake(uint32) (int, error) {
	sw.s.Signal(1)
	return 1, nil
}

func (sw *semaTimedWaiter) wait(unused uint32, timeout time.Duration) error {
	if !sw.s.WaitTimeout(timeout) {
		return os.NewSyscallError("SEMTIMEDOP", unix.EAGAIN)
	}
	return nil
}
