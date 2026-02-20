// backend/logger/handler_file.go

package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type FileLogHandler struct {
	level  slog.Level
	writer io.Writer
	mu     sync.Mutex
	attrs  []slog.Attr
	groups []string
}

func NewFileLogHandler(level slog.Level, writer io.Writer) *FileLogHandler {
	if writer == nil {
		writer = io.Discard
	}
	return &FileLogHandler{
		level:  level,
		writer: writer,
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

func (h *FileLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *FileLogHandler) Handle(_ context.Context, rec slog.Record) error {
	timestamp := rec.Time.Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("%s [%s] %s", timestamp, rec.Level, rec.Message)

	// Atributos acumulados
	for _, attr := range h.attrs {
		logLine += fmt.Sprintf(" %s=%v", h.formatKey(attr.Key), attr.Value.Any())
	}

	// Atributos del record
	rec.Attrs(func(attr slog.Attr) bool {
		logLine += fmt.Sprintf(" %s=%v", h.formatKey(attr.Key), attr.Value.Any())
		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := fmt.Fprintln(h.writer, logLine)
	return err
}

func (h *FileLogHandler) formatKey(key string) string {
	for _, group := range h.groups {
		key = group + "." + key
	}
	return key
}

func (h *FileLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &FileLogHandler{
		level:  h.level,
		writer: h.writer,
		attrs:  newAttrs,
		groups: append([]string{}, h.groups...),
	}
}

func (h *FileLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &FileLogHandler{
		level:  h.level,
		writer: h.writer,
		attrs:  append([]slog.Attr{}, h.attrs...),
		groups: append(append([]string{}, h.groups...), name),
	}
}
