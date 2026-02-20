// backend/logger/handler_console.go

package logger

import (
	"context"
	"fmt"
	"log/slog"
)

type ConsoleLogHandler struct {
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

func NewConsoleLogHandler(level slog.Level) *ConsoleLogHandler {
	return &ConsoleLogHandler{
		level:  level,
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

func (h *ConsoleLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ConsoleLogHandler) Handle(_ context.Context, rec slog.Record) error {
	timestamp := rec.Time.Format("15:04:05")
	level := rec.Level.String()

	// Find module in accumulated attributes
	modulePrefix := ""
	for _, attr := range h.attrs {
		if attr.Key == LoggerAttributeKeyModule {
			modulePrefix = attr.Value.String()
			break
		}
	}

	// If not found in accumulated, search in the record
	if modulePrefix == "" {
		rec.Attrs(func(attr slog.Attr) bool {
			if attr.Key == LoggerAttributeKeyModule {
				modulePrefix = attr.Value.String()
				return false
			}
			return true
		})
	}

	// Build the message with attributes
	message := rec.Message

	// Add attributes from the record
	rec.Attrs(func(attr slog.Attr) bool {
		if attr.Key != LoggerAttributeKeyModule { // Avoid duplicating the module
			message += fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any())
		}
		return true
	})

	// Add accumulated attributes (except module)
	for _, attr := range h.attrs {
		if attr.Key != LoggerAttributeKeyModule {
			message += fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any())
		}
	}

	// Print with the desired format
	fmt.Printf("[%s] %s [%s]: %s\n", timestamp, level, modulePrefix, message)
	return nil
}

func (h *ConsoleLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &ConsoleLogHandler{
		level:  h.level,
		attrs:  newAttrs,
		groups: append([]string{}, h.groups...),
	}
}

func (h *ConsoleLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &ConsoleLogHandler{
		level:  h.level,
		attrs:  append([]slog.Attr{}, h.attrs...),
		groups: append(append([]string{}, h.groups...), name),
	}
}
