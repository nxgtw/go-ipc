// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build ignore

package sync

import "time"

type semaWaiter struct {
	s *semaphore
}

func newSemaWaiter(s Semaphore) *semaWaiter {
	return &semaWaiter{s: s.(*semaphore)}
}

func (sw *semaWaiter) wake(count uint32) (int, error) {
	sw.s.Signal(int(count))
	return int(count), nil
}

func (sw *semaWaiter) wait(unused uint32, timeout time.Duration) error {
	sw.s.Wait()
	return nil
}
