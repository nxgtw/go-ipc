// Copyright 2016 Aleksandr Demakin. All rights reserved.

// +build darwin freebsd

package sync

import (
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/nxgtw/go-ipc/internal/common"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// This is the emulation of semtimedop.
// As darwin/bsd don't have semtimedop, we call semop
// waiting for timeout to elapse in another goroutine.
// After that, it sends SIGUSR2 to the blocked goroutie,
// forcing it to interrupt the syscall with EINTR.
// This, however, has some side effects:
//	- it locks the thread to get valid id.
//	- SIGUSR2 won't be ignored for the calling thread, if it was before.
// The code uses the same idea, as used here:
// https://github.com/attie/libxbee3/blob/master/xsys_darwin/sem_timedwait.c

type threadInterrupter struct {
	state int32
}

func (ti *threadInterrupter) start(timeout time.Duration) error {
	// first, get thread id. goroutine must be locked on the thread.
	tid, err := gettid()
	if err != nil {
		return errors.Wrap(err, "failed to get thread id")
	}
	// then, restore SIGUSR2 handler if it was ignored before.
	// we don't know, if it was really igored, so we do it unconditionally.
	// side effect: SIGUSR2 won't be ignored again after the operation is complete.
	signal.Notify(make(chan os.Signal, 1), unix.SIGUSR2)
	signal.Reset(unix.SIGUSR2)
	go func() {
		time.Sleep(timeout)
		if atomic.LoadInt32(&ti.state) == 0 {
			killThread(tid)
		}
	}()
	return nil
}

func (ti *threadInterrupter) done() {
	atomic.StoreInt32(&ti.state, 1)
}

func doSemaTimedWait(id int, timeout time.Duration) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ti := threadInterrupter{}
	b := sembuf{semnum: 0, semop: int16(-1), semflg: 0}
	if err := ti.start(timeout); err != nil {
		panic(errors.Wrap(err, "failed to setup timeout"))
	}
	err := semop(id, []sembuf{b})
	ti.done()
	if err == nil {
		return true
	}
	if common.IsInterruptedSyscallErr(err) {
		return false
	}
	panic(err)
}
