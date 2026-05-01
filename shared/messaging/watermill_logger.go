package messaging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/fms/fms/shared/logger"
)

// WatermillLoggerAdapter adapts shared/logger.Logger to watermill.LoggerAdapter
type WatermillLoggerAdapter struct {
	logger *logger.Logger
}

func NewWatermillLoggerAdapter(l *logger.Logger) *WatermillLoggerAdapter {
	return &WatermillLoggerAdapter{logger: l}
}

func (l *WatermillLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	args := make([]any, 0, len(fields)+2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	args = append(args, "error", err)
	l.logger.Error(msg, args...)
}

func (l *WatermillLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.logger.Info(msg, args...)
}

func (l *WatermillLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.logger.Debug(msg, args...)
}

func (l *WatermillLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.logger.Debug(msg, args...)
}

func (l *WatermillLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return l
}
