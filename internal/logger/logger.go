package logger

import (
	"fmt"
)

type Logger struct{}

func (l *Logger) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}
