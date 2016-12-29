// Copyright 2015 Aleksandr Demakin. All rights reserved.

package sync

import (
	"os"
	"time"

	"github.com/pkg/errors"
)

// ensureOpenFlags ensures, that no other flags but os.O_CREATE and os.O_EXCL are set.
func ensureOpenFlags(flags int) error {
	if flags & ^(os.O_CREATE|os.O_EXCL) != 0 {
		return errors.New("only os.O_CREATE and os.O_EXCL are allowed")
	}
	return nil
}

// waitWaker is an object, which implements wake/wait semantics.
type waitWaker interface {
	wake(count int32) (int, error)
	wait(value int32, timeout time.Duration) error
}
