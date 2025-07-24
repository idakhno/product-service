package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// Logger defines a common interface for logging.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	WithTrace(ctx context.Context) Logger
}

// SlogAdapter implements the Logger interface using the standard library's slog.
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new logger adapter based on the environment.
func NewSlogAdapter(env string) Logger {
	var handler slog.Handler

	switch env {
	case "local":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "dev":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "prod":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	log := slog.New(handler)

	return &SlogAdapter{logger: log}
}

func (s *SlogAdapter) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

func (s *SlogAdapter) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

func (s *SlogAdapter) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

func (s *SlogAdapter) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

func (s *SlogAdapter) WithTrace(ctx context.Context) Logger {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return &SlogAdapter{
			logger: s.logger.With("trace_id", span.SpanContext().TraceID().String()),
		}
	}
	return s
}
