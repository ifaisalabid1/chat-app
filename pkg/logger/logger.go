package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	*slog.Logger
}

func New(env string) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		AddSource: env == "development",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format(time.RFC3339))
			}
			return a
		},
	}

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{slog.New(handler)}
}

func (l *Logger) Error(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		attrs = append(attrs, slog.String("source", file))
		attrs = append(attrs, slog.Int("line", line))
	}

	l.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	l.Error(ctx, msg, err, attrs...)
	os.Exit(1)
}
