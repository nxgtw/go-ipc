// Copyright 2015 Aleksandr Demakin. All rights reserved.

// +build linux,amd64

package sync

import (
	"time"

	"bitbucket.org/avd/go-ipc/internal/common"
)

func (m *mutex) LockTimeout(timeout time.Duration) bool {
	for {
		b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
		opStart := time.Now()
		err := semtimedop(m.id, []sembuf{b}, common.TimeoutToTimeSpec(timeout))
		if err == nil {
			return true
		}
		if isTimeoutErr(err) {
			return false
		}
		if !isInterruptedSyscallErr(err) {
			panic(err)
		}
		if timeout >= 0 {
			// we were interrupted by a signal. recalculate timeout
			elapsed := time.Now().Sub(opStart)
			if timeout > elapsed {
				timeout = timeout - elapsed
			} else {
				return false
			}
		}
	}
}
