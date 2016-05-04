// Copyright 2016 Aleksandr Demakin. All rights reserved.

package sync

import "unsafe"

const (
	cFUTEX_WAIT        = 0
	cFUTEX_WAKE        = 1
	cFUTEX_REQUEUE     = 3
	cFUTEX_CMP_REQUEUE = 4
	cFUTEX_WAKE_OP     = 5

	cFUTEX_PRIVATE_FLAG   = 128
	cFUTEX_CLOCK_REALTIME = 256
)

func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) (int32, error) {
	return 0, nil
}
