// Copyright 2016 Aleksandr Demakin. All rights reserved.

package ipc_testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFilesFromOutput(t *testing.T) {
	type td struct {
		in  string
		out []string
	}
	a := assert.New(t)
	data := []td{
		{
			in:  "[a.go]",
			out: []string{"a.go"},
		},
		{
			in:  " [ a.go  b.go] ",
			out: []string{"a.go", "b.go"},
		},
		{
			in:  " [ a.go  b.go c .go] ",
			out: []string{"a.go", "b.go", "c.go"},
		},
		{
			in:  " [ a.go  b.go c d .go] ",
			out: []string{"a.go", "b.go", "cd.go"},
		},
		{
			in:  " [ a.go  b.go c d ] ",
			out: []string{"a.go", "b.go", "cd"},
		},
		{
			in:  " [ a b c d ] ",
			out: []string{"abcd"},
		},
	}
	for _, d := range data {
		a.Equal(d.out, buildFilesFromOutput(d.in))
	}
}
