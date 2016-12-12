// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package sync

import "time"

type semaWaiter struct {
	s Semaphore
}

func newSemaWaiter(s Semaphore) *semaWaiter {
	return &semaWaiter{s: s}
}

func (sw *semaWaiter) wake() {
	sw.s.Signal(1)
}

func (sw *semaWaiter) wait(timeout time.Duration) error {
	sw.s.Wait()
	return nil
}
