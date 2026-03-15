package log

import (
	"fmt"
	"io"
	stdlog "log"
)

type Fields map[string]any

type Logger interface {
	Trace(message string, fields Fields)
	Debug(message string, fields Fields)
	Info(message string, fields Fields)
	Warn(message string, fields Fields)
	Error(err error, fields Fields)
}

type nopLogger struct{}

func Nop() Logger {
	return nopLogger{}
}

func (nopLogger) Trace(string, Fields) {}

func (nopLogger) Debug(string, Fields) {}

func (nopLogger) Info(string, Fields) {}

func (nopLogger) Warn(string, Fields) {}

func (nopLogger) Error(error, Fields) {}

type StdLogger struct {
	logger *stdlog.Logger
}

func NewStdLogger(logger *stdlog.Logger) Logger {
	if logger == nil {
		logger = stdlog.New(io.Discard, "", 0)
	}
	return &StdLogger{logger: logger}
}

func (l *StdLogger) Trace(message string, fields Fields) {
	l.printf("TRACE", message, fields)
}

func (l *StdLogger) Debug(message string, fields Fields) {
	l.printf("DEBUG", message, fields)
}

func (l *StdLogger) Info(message string, fields Fields) {
	l.printf("INFO", message, fields)
}

func (l *StdLogger) Warn(message string, fields Fields) {
	l.printf("WARN", message, fields)
}

func (l *StdLogger) Error(err error, fields Fields) {
	message := ""
	if err != nil {
		message = err.Error()
	}
	l.printf("ERROR", message, fields)
}

func (l *StdLogger) printf(level string, message string, fields Fields) {
	if len(fields) == 0 {
		l.logger.Printf("[%s] %s", level, message)
		return
	}
	l.logger.Printf("[%s] %s %v", level, message, map[string]any(fields))
}

func FormatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}
	return fmt.Sprint(map[string]any(fields))
}
