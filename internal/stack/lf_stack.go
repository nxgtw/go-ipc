// Copyright 2017 Aleksandr Demakin. All rights reserved.

package stack

import (
	"unsafe"
)

type lfsHead struct {
	idx uint16
	ver uint16
}

type lfs struct {
	cap      int32
	elemsize int32
	head     lfsHead
}

type LfStack struct {
	ptr *lfs
}

func NewLfStack(raw unsafe.Pointer) *LfStack {
	return &LfStack{}
}
