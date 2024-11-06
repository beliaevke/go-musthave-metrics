package main

import (
	"testing"
)

func mulfunc(i int) (int, error) {
	return i * 2, nil
}

func TestFunc(t *testing.T) {
	var i int
	myfunc := func() error {
		return nil
	}
	myfunc()
	if true {
		i := 7            //nolint
		i, _ = mulfunc(i) //nolint
	}
	i, _ = i+1, myfunc() //nolint
}
