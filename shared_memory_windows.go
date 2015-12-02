// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import "os"

type memoryObjectImpl struct {
	file *os.File
}

type memoryRegionImpl struct {
	data       []byte
	size       int
	pageOffset int64
}
