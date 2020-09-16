package context

import (
	"fmt"
	"os"
)

type Logger interface {
	Errorf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type DefaultLogger struct {
}

func (l DefaultLogger) Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func (l DefaultLogger) Infof(format string, a ...interface{}) {
	// no op
}

func (l DefaultLogger) Debugf(format string, a ...interface{}) {
	// no op
}
