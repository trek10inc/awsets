package context

import (
	"fmt"
	"os"
)

type Logger interface {
	Errorf(format string, a ...interface{})
	Errorln(a ...interface{})
	Infof(format string, a ...interface{})
	Infoln(a ...interface{})
}

type DefaultLogger struct {
}

func (l DefaultLogger) Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func (l DefaultLogger) Errorln(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func (l DefaultLogger) Infof(format string, a ...interface{}) {
	// no op
}

func (l DefaultLogger) Infoln(a ...interface{}) {
	// no op
}
