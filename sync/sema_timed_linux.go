// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

// AddTimeout add the given value to the semaphore's value.
// If the operation locks, it waits for not more, than timeout.
// This call is supported on linux only.
func (s *Semaphore) AddTimeout(value int, timeout time.Duration) error {
	f := func(curTimeout time.Duration) error {
		b := sembuf{semnum: 0, semop: int16(value), semflg: 0}
		return semtimedop(s.id, []sembuf{b}, common.TimeoutToTimeSpec(curTimeout))
	}
	return common.UninterruptedSyscallTimeout(f, timeout)
}

// LockTimeout tries to lock the locker, waiting for not more, than timeout.
// This call is supported on linux only.
func (m *SemaMutex) LockTimeout(timeout time.Duration) bool {
	err := m.inplace.lockTimeout(timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}

func (m *SemaMutex) wait(ptr *uint32, timeout time.Duration) error {
	return m.s.AddTimeout(-1, timeout)
}
