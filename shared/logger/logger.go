package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

type Logger struct {
	logger *slog.Logger
	attrs  []slog.Attr
	mu     sync.RWMutex
}

var defaultLogger *Logger
var once sync.Once

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Option func(*slog.HandlerOptions) *slog.HandlerOptions

func WithLevel(level Level) Option {
	return func(opts *slog.HandlerOptions) *slog.HandlerOptions {
		switch level {
		case LevelDebug:
			opts.Level = slog.LevelDebug
		case LevelInfo:
			opts.Level = slog.LevelInfo
		case LevelWarn:
			opts.Level = slog.LevelWarn
		case LevelError:
			opts.Level = slog.LevelError
		}
		return opts
	}
}

func WithJSON() Option {
	return func(opts *slog.HandlerOptions) *slog.HandlerOptions {
		return opts
	}
}

func WithDebugSampling(rate float64) Option {
	return func(opts *slog.HandlerOptions) *slog.HandlerOptions {
		opts.Level = slog.LevelDebug
		return opts
	}
}

func New(opts ...Option) *Logger {
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	for _, opt := range opts {
		handlerOpts = opt(handlerOpts)
	}

	l := &Logger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, handlerOpts)),
	}

	return l
}

func Default() *Logger {
	once.Do(func() {
		defaultLogger = New()
	})
	return defaultLogger
}

func (l *Logger) With(key string, value any) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	clone := &Logger{
		logger: l.logger,
		attrs:  append([]slog.Attr{}, l.attrs...),
	}
	clone.attrs = append(clone.attrs, slog.Any(key, value))
	return clone
}

func (l *Logger) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Error(msg, args...)
}

func (l *Logger) Log(ctx context.Context, level Level, msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	}

	l.logger.Log(ctx, slogLevel, msg, args...)
}