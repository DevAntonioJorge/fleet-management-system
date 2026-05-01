package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	l := New(WithJSON())
	if l == nil {
		t.Error("Logger should not be nil")
	}
}

func TestLoggerWith(t *testing.T) {
	l := New()
	l2 := l.With("key", "value")
	if l2 == nil {
		t.Error("Logger with context should not be nil")
	}
}

func TestDefaultLogger(t *testing.T) {
	l := Default()
	if l == nil {
		t.Error("Default logger should not be nil")
	}
}