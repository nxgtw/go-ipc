// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package sync

import "time"

func (m *SemaMutex) wait(ptr *uint32, timeout time.Duration) error {
	return m.s.Add(-1)
}
