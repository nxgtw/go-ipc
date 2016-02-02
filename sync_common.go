// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

func checkMutexOpenMode(mode int) bool {
	return /*mode == O_OPEN_OR_CREATE ||*/ mode == O_CREATE_ONLY || mode == O_OPEN_ONLY
}
