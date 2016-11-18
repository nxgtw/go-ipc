// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package sync

import "time"

type semaWaiter struct {
	s *Semaphore
}

func newSemaWaiter(s *Semaphore) *semaWaiter {
	return &semaWaiter{s: s}
}

func (sw *semaWaiter) set(*uint32) {

}

func (sw *semaWaiter) wake() {
	if err := sw.s.Add(1); err != nil {
		panic(err)
	}
}

func (sw *semaWaiter) wait(timeout time.Duration) error {
	return sw.s.Add(-1)
}
