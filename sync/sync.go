// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// IPCLocker is a minimal interface, which must be satisfied by any synchronization primitive
// on any platform.
type IPCLocker interface {
	sync.Locker
	io.Closer
}

// TimedIPCLocker is a locker, whose lock operation can be limited with duration.
type TimedIPCLocker interface {
	IPCLocker
	// LockTimeout tries to lock the locker, waiting for not more, than timeout
	LockTimeout(timeout time.Duration) bool
}

// ensureOpenFlags ensures, that no other flags but os.O_CREATE and os.O_EXCL are set.
func ensureOpenFlags(flags int) error {
	if flags & ^(os.O_CREATE|os.O_EXCL) != 0 {
		return errors.New("only os.O_CREATE and os.O_EXCL are allowed")
	}
	return nil
}
