package slogdiscard

import (
	"context"
	"log/slog"
)

type DiscardLogger struct{}

func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

type DiscardHandler struct{}

func (h *DiscardHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (h *DiscardHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (h *DiscardHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *DiscardHandler) WithGroup(name string) slog.Handler {
	return h
}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}
