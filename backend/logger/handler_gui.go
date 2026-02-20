// backend/logger/handler_gui.go

package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	//"os"
)

type GuiLogHandler struct {
	level             slog.Level
	addLogMessageFunc func(string)
	attrs             []slog.Attr
	groups            []string
}

func NewGuiLogHandler(level slog.Level, addLogMessageFunc func(string)) *GuiLogHandler {
	/*
		if addLogMessageFunc == nil {
			fmt.Fprintln(os.Stderr, "WARNING: NewGuiLogHandler received nil log function.")
			return &GuiLogHandler{
				level:             level,
				addLogMessageFunc: nil,
				attrs:             make([]slog.Attr, 0),
				groups:            make([]string, 0),
			}
		}
	*/
	return &GuiLogHandler{
		level:             level,
		addLogMessageFunc: addLogMessageFunc,
		attrs:             make([]slog.Attr, 0),
		groups:            make([]string, 0),
	}
}

func (h *GuiLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *GuiLogHandler) Handle(_ context.Context, rec slog.Record) error {
	if h.addLogMessageFunc == nil {
		return nil
	}

	var builder strings.Builder

	// Format timestamp and main message.
	builder.WriteString(fmt.Sprintf("[%s]: %s", rec.Time.Format("15:04:05"), rec.Message))

	// Iterate through all attributes in the record. The slog package automatically
	// includes attributes from `With` calls in the record itself.
	rec.Attrs(func(attr slog.Attr) bool {
		// Exclude the module attribute from the GUI log, as requested.
		if attr.Key != LoggerAttributeKeyModule {
			// Append other attributes safely using the .String() method.
			builder.WriteString(fmt.Sprintf(" %s=%s", attr.Key, attr.Value.String()))
		}
		return true // Continue iterating.
	})

	// Send the final, single string to the GUI function.
	h.addLogMessageFunc(builder.String())

	return nil
}

func (h *GuiLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &GuiLogHandler{
		level:             h.level,
		addLogMessageFunc: h.addLogMessageFunc,
		attrs:             newAttrs,
		groups:            append([]string{}, h.groups...),
	}
}

func (h *GuiLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &GuiLogHandler{
		level:             h.level,
		addLogMessageFunc: h.addLogMessageFunc,
		attrs:             append([]slog.Attr{}, h.attrs...),
		groups:            append(append([]string{}, h.groups...), name),
	}
}

