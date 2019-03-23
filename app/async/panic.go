package async

import (
	"fmt"
	"runtime"
	"strings"
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
	close(ch)
}

func NewWrappedPanic(v interface{}) *WrappedPanic {
	var buf [16384]byte
	stack := string(buf[0:runtime.Stack(buf[:], false)])
	return &WrappedPanic{v, chopStack(stack, "panic(")}
}

func chopStack(s, panicText string) string {
	lines := strings.Split(s, "\n")

	for idx := 1; idx < len(lines); idx++ {
		line := lines[idx]
		if strings.HasPrefix(line, panicText) {
			idx += 2 // omit the part of the stack containing `recover` and `panic`
			return fmt.Sprintf("%s\n%s", lines[0], strings.Join(lines[idx:], "\n"))
		}
	}

	return s
}
