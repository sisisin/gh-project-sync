package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sisisin/gh-project-sync/lib/appcontext"
)

type LogHandler struct {
	slog.Handler
}

var _ slog.Handler = (*LogHandler)(nil)

func (h *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if val := appcontext.GetTraceID(ctx); val != "" {
		projectID := appcontext.GetProjectID(ctx)
		key := fmt.Sprintf("projects/%s/traces/%s", projectID, val)
		r.AddAttrs(slog.Attr{Key: "trace", Value: slog.StringValue(key)})
	}

	return h.Handler.Handle(ctx, r)
}

func SetDefaultLogger() {
	logger := slog.New(&LogHandler{slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.MessageKey:
				a = slog.Attr{
					Key:   "message",
					Value: a.Value,
				}
			case slog.LevelKey:
				a = slog.Attr{
					Key:   "severity",
					Value: a.Value,
				}
			case slog.SourceKey:
				a = slog.Attr{
					Key:   "logging.googleapis.com/sourceLocation",
					Value: a.Value,
				}
			}
			return a
		},
	})})
	slog.SetDefault(logger)
}

func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}
func Infof(ctx context.Context, format string, args ...any) {
	slog.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}
