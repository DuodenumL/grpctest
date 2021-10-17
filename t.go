package main

import "fmt"

type T struct {
}

func (t *T) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
