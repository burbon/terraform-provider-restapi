package log

import (
	"fmt"
	stdlog "log"
	"os"
)

// Logger encapsulates std Logger and adds debug logs.
type Logger struct {
	logger *stdlog.Logger
	debug  bool
}

// New creates new Logger
func New(debug bool) *Logger {
	return &Logger{
		// logger: stdlog.New(os.Stderr, "", stdlog.LstdFlags),
		logger: stdlog.New(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile),
		debug:  debug,
	}
}

// Printf prints using std logger.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
}

// Debugf prints using std logger only if debug flag is set.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.debug {
		l.logger.Output(2, fmt.Sprintf(format, v...))
	}
}
