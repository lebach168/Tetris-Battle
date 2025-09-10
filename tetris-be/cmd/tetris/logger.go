package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type MessageOnlyHandler struct {
	io.Writer
	level slog.Level
}

func (h *MessageOnlyHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level
}

func (h *MessageOnlyHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format("2006/01/02 15:04:05")
	fmt.Fprintf(h.Writer, "%s %s %s\n", ts, r.Level, r.Message)
	return nil
}

func (h *MessageOnlyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MessageOnlyHandler) WithGroup(_ string) slog.Handler {
	return h
}
func NewLogger() *slog.Logger {
	h := &MessageOnlyHandler{
		Writer: os.Stdout,
		level:  slog.LevelInfo,
	}
	return slog.New(h)
}
