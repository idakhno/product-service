package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// Logger defines the interface for logging in the application.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	WithTrace(ctx context.Context) Logger // Creates a new logger with trace ID from context
}

// SlogAdapter is an adapter for the standard slog.Logger.
// Implements the Logger interface for consistent logging across the application.
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new logger adapter based on the environment.
// For local environment uses text format with Debug level.
// For dev and prod environments uses JSON format (Debug for dev, Info for prod).
func NewSlogAdapter(env string) Logger {
	var handler slog.Handler

	switch env {
	case "local":
		// Text format for development convenience
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "dev":
		// JSON format for dev environment
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "prod":
		// JSON format for production, Info level and above only
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		// Default to text format with Debug level
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	log := slog.New(handler)

	return &SlogAdapter{logger: log}
}

// Info logs an informational message.
func (s *SlogAdapter) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (s *SlogAdapter) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

// Error logs an error message.
func (s *SlogAdapter) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

// Debug logs a debug message.
func (s *SlogAdapter) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

// WithTrace creates a new logger with trace ID from OpenTelemetry context.
// Returns the original logger if trace ID is not present.
func (s *SlogAdapter) WithTrace(ctx context.Context) Logger {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return &SlogAdapter{
			logger: s.logger.With("trace_id", span.SpanContext().TraceID().String()),
		}
	}
	return s
}
