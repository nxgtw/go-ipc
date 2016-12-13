// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd windows

package sync

import "time"

type semaWaiter struct {
	s Semaphore
}

func newSemaWaiter(s Semaphore) *semaWaiter {
	return &semaWaiter{s: s}
}

func (sw *semaWaiter) wake(uint32) (int, error) {
	sw.s.Signal(1)
	return 1, nil
}

func (sw *semaWaiter) wait(unused uint32, timeout time.Duration) error {
	sw.s.Wait()
	return nil
}
