// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"github.com/nxgtw/go-ipc/internal/common"
)

const (
	// CSemMaxVal is the maximum semaphore value,
	// which is guaranteed to be supported on all platforms.
	CSemMaxVal = 32767
)

// Semaphore is a synchronization object with a resource counter,
// which can be used to control access to a shared resource.
// It provides access to actual OS semaphore primitive via:
//	CreateSemaprore on windows
//	semget on unix
type Semaphore semaphore

// NewSemaphore creates new semaphore with the given name.
//	name - object name.
//	flag - flag is a combination of open flags from 'os' package.
//	perm - object's permission bits.
//	initial - this value will be added to the semaphore's value, if it was created.
func NewSemaphore(name string, flag int, perm os.FileMode, initial int) (*Semaphore, error) {
	result, err := newSemaphore(name, flag, perm, initial)
	if err != nil {
		return nil, err
	}
	return (*Semaphore)(result), nil
}

// Signal increments the value of semaphore variable by 1, waking waiting process (if any).
func (s *Semaphore) Signal(count int) {
	(*semaphore)(s).signal(count)
}

// Wait decrements the value of semaphore variable by -1, and blocks if the value becomes negative.
func (s *Semaphore) Wait() {
	(*semaphore)(s).wait()
}

// Close closes the semaphore.
func (s *Semaphore) Close() error {
	return (*semaphore)(s).close()
}

// WaitTimeout decrements the value of semaphore variable by 1.
// If the value becomes negative, it waites for not longer than timeout.
// On darwin and freebsd this func has some side effects, see sema_timed_bsd.go for details.
func (s *Semaphore) WaitTimeout(timeout time.Duration) bool {
	return (*semaphore)(s).waitTimeout(timeout)
}

// DestroySemaphore removes the semaphore permanently.
func DestroySemaphore(name string) error {
	return destroySemaphore(name)
}

type semaWaiter struct {
	s *Semaphore
}

func newSemaWaiter(s *Semaphore) *semaWaiter {
	return &semaWaiter{s: s}
}

func (sw *semaWaiter) wake(count int32) (int, error) {
	sw.s.Signal(int(count))
	return int(count), nil
}

func (sw *semaWaiter) wait(unused int32, timeout time.Duration) error {
	if !sw.s.WaitTimeout(timeout) {
		return common.NewTimeoutError("SEMWAWIT")
	}
	return nil
}
