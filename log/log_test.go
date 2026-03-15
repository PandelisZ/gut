package log

import (
	"bytes"
	"errors"
	stdlog "log"
	"strings"
	"testing"
)

func TestNopLoggerDoesNotPanic(t *testing.T) {
	logger := Nop()
	logger.Trace("trace", nil)
	logger.Debug("debug", Fields{"key": "value"})
	logger.Info("info", nil)
	logger.Warn("warn", nil)
	logger.Error(errors.New("boom"), nil)
}

func TestStdLoggerWritesMessages(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStdLogger(stdlog.New(&buf, "", 0))
	logger.Info("hello", Fields{"answer": 42})
	logger.Error(errors.New("boom"), nil)

	output := buf.String()
	if !strings.Contains(output, "[INFO] hello") {
		t.Fatalf("expected info log output, got %q", output)
	}
	if !strings.Contains(output, "answer") {
		t.Fatalf("expected structured field output, got %q", output)
	}
	if !strings.Contains(output, "[ERROR] boom") {
		t.Fatalf("expected error log output, got %q", output)
	}
}
