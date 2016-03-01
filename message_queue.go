// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc

import "os"

func checkMqPerm(perm os.FileMode) bool {
	return uint(perm)&0111 == 0
}
