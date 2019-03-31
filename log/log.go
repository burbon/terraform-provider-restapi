package log

import (
	"fmt"
	stdlog "log"
	"os"
)

type Logger struct {
	logger *stdlog.Logger
	debug  bool
}

func New(debug bool) *Logger {
	return &Logger{
		// logger: stdlog.New(os.Stderr, "", stdlog.LstdFlags),
		logger: stdlog.New(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile),
		debug:  debug,
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	if l.debug {
		l.logger.Output(2, fmt.Sprintf(format, v...))
	}
}
