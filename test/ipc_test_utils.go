// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc_test

import (
	"bytes"
	"fmt"
	"io"
)

// StringToBytes takes an input string in a 2-hex-symbol per byte format
// and returns corresponding byte array.
// Input must not contain any symbols except [A-F0-9]
func StringToBytes(input string) ([]byte, error) {
	if len(input)%2 != 0 {
		return nil, fmt.Errorf("invalid byte array len")
	}
	var err error
	var b byte
	buff := bytes.NewBuffer(nil)
	for err == nil {
		if len(input) < 2 {
			err = io.EOF
		} else if _, err = fmt.Sscanf(input[:2], "%X", &b); err == nil {
			buff.WriteByte(b)
			input = input[2:]
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buff.Bytes(), nil
}

// BytesToString convert a byte slice into its string representation.
// Each byte is represented as a 2 upper-case letters for A-F
func BytesToString(data []byte) string {
	buff := bytes.NewBuffer(nil)
	for _, value := range data {
		if value < 16 { // force leading 0 for 1-digit values
			buff.WriteString("0")
		}
		buff.WriteString(fmt.Sprintf("%X", value))
	}
	return buff.String()
}
