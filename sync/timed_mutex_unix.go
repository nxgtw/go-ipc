// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux,amd64

package sync

import (
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

func (m *mutex) LockTimeout(timeout time.Duration) bool {
	f := func(curTimeout time.Duration) error {
		b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
		return semtimedop(m.id, []sembuf{b}, common.TimeoutToTimeSpec(curTimeout))
	}
	err := common.UninterruptedSyscallTimeout(f, timeout)
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}
