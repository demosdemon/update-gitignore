package app

import (
	"bytes"
	"fmt"
	"runtime"
)

type WrappedPanic struct {
	Value interface{}
	Stack string
}

func (p *WrappedPanic) Error() string {
	return fmt.Sprintf("panic: %v\n%v", p.Value, p.Stack)
}

func RecoverFromPanic(ch chan<- error) {
	if p := recover(); p != nil {
		ch <- NewWrappedPanic(p)
	}
}

func NewWrappedPanic(v interface{}) *WrappedPanic {
	var buf [16384]byte
	stack := buf[0:runtime.Stack(buf[:], false)]
	return &WrappedPanic{v, chopStack(stack, "panic(")}
}

func chopStack(s []byte, panicText string) string {
	lfFirst := bytes.IndexByte(s, '\n')
	if lfFirst == -1 {
		return string(s)
	}

	stack := s[lfFirst:]
	f := []byte(panicText)
	panicLine := bytes.Index(stack, f)
	if panicLine == -1 {
		return string(s)
	}

	stack = stack[panicLine+1:]
	for i := 0; i < 2; i++ {
		nextLine := bytes.IndexByte(stack, '\n')
		if nextLine == -1 {
			return string(s)
		}
		stack = stack[nextLine+1:]
	}

	return string(s[:lfFirst+1]) + string(stack)
}
