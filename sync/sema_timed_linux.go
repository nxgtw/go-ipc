// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"time"

	"github.com/nxgtw/go-ipc/internal/common"
)

func doSemaTimedWait(id int, timeout time.Duration) bool {
	err := common.UninterruptedSyscallTimeout(func(curTimeout time.Duration) error {
		b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
		return semtimedop(id, []sembuf{b}, common.TimeoutToTimeSpec(curTimeout))
	}, timeout)
	if err == nil {
		return true
	}
	if common.IsTimeoutErr(err) {
		return false
	}
	panic(err)
}
