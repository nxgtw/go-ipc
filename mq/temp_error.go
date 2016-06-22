// Copyright 2016 Aleksandr Demakin. All rights reserved.

package mq

type temporaryError struct {
	inner error
}

func (e *temporaryError) isTemporary() bool {
	return true
}

func (e *temporaryError) Error() string {
	return e.inner.Error()
}

func newTemporaryError(inner error) *temporaryError {
	return &temporaryError{inner: inner}
}

func isTemporaryError(e error) bool {
	if tmp, ok := e.(*temporaryError); ok {
		return tmp.isTemporary()
	}
	return false
}
