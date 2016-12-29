// Copyright 2016 Aleksandr Demakin. All rights reserved.

// Package fifo implements first-in-first-out objects logic.
// It gives access to OS-native FIFO objects via:
//	CreateNamedPipe on windows
//	Mkfifo on unix
package fifo
