// Copyright 2015 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	objName = flag.String("object", "", "mq name")
)

const usage = `  test program for message queues.
available commands:
  create {max_size} {max_msg_len}
  destroy
  test {expected values byte array}
  write {values byte array}
byte array should be passed as a continuous string of 2-symbol hex byte values like '01020A'
`

func runCommand() error {
	return nil
}

func main() {
	flag.Parse()
	if len(*objName) == 0 || flag.NArg() == 0 {
		fmt.Print(usage)
		flag.Usage()
		os.Exit(1)
	}
	if err := runCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
