// backend/logger/logger.go

package logger

import (
	"io"
	"log/slog"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

const LoggerAttributeKeyModule = "module"

type Logger struct {
	console     *slog.Logger
	gui         *slog.Logger
	file        *slog.Logger
	fileCloser  io.Closer
	closeOnce   sync.Once
	mu          sync.Mutex
	level       slog.Level
	logFilePath string
}

const (
	defaultLogMaxSizeMB  = 10
	defaultLogMaxBackups = 3
	defaultLogMaxAgeDays = 7
	defaultLogCompress   = true
)

// NewLogger creates a new Logger instance
func NewLogger(
	addGuiLogMessageFunc func(string),
	levelConsole slog.Level,
	levelGui slog.Level,
	levelFile slog.Level,
	logFilePath string,
) (*Logger, error) {
	// Console Logger
	consoleHandler := NewConsoleLogHandler(levelConsole)
	consoleLogger := slog.New(consoleHandler)

	// GUI Logger
	guiHandler := NewGuiLogHandler(levelGui, addGuiLogMessageFunc)
	guiLogger := slog.New(guiHandler)

	// File Logger Setup
	var fileLogger *slog.Logger
	var actualFileCloser io.Closer
	fileWriter := io.Discard

	if logFilePath != "" {
		lj := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    defaultLogMaxSizeMB,
			MaxBackups: defaultLogMaxBackups,
			MaxAge:     defaultLogMaxAgeDays,
			Compress:   defaultLogCompress,
			LocalTime:  true,
		}

		if _, err := lj.Write([]byte("File logging initialized.\n")); err != nil {
			consoleLogger.Error("Failed to initialize file logging", "path", logFilePath, "error", err)
		} else {
			fileWriter = lj
			actualFileCloser = lj
		}
	}

	fileHandler := NewFileLogHandler(levelFile, fileWriter)
	fileLogger = slog.New(fileHandler)

	logger := &Logger{
		console:     consoleLogger,
		gui:         guiLogger,
		file:        fileLogger,
		fileCloser:  actualFileCloser,
		level:       slog.LevelInfo,
		logFilePath: logFilePath,
	}

	return logger, nil
}

// Close closes the logger and its resources
func (l *Logger) Close() error {
	var err error
	l.closeOnce.Do(func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if l.fileCloser == nil {
			return
		}

		err = l.fileCloser.Close()
		l.fileCloser = nil
	})

	return err
}

// ============================================================================
// STANDARD LOGGING METHODS (Console + File)
// ============================================================================
// These are the default methods - they log to Console and File only.
// Use the *Gui() variants below for messages that should also appear in GUI.

func (l *Logger) Info(msg string, args ...any) {
	l.console.Info(msg, args...)
	l.file.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.console.Error(msg, args...)
	l.file.Error(msg, args...)
}
func (l *Logger) Warn(msg string, args ...any) {
	l.console.Warn(msg, args...)
	l.file.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.console.Debug(msg, args...)
	l.file.Debug(msg, args...)
}

// ============================================================================
// GUI-INCLUSIVE LOGGING METHODS (Console + File + GUI)
// ============================================================================
// Use these for important user-facing messages that should appear in the GUI.
// Examples: "Server started", "OBS connected", "Critical errors", etc.

func (l *Logger) InfoGui(msg string, args ...any) {
	l.console.Info(msg, args...)
	l.file.Info(msg, args...)
	l.gui.Info(msg, args...)
}

func (l *Logger) ErrorGui(msg string, args ...any) {
	l.console.Error(msg, args...)
	l.file.Error(msg, args...)
	l.gui.Error(msg, args...)
}

func (l *Logger) WarnGui(msg string, args ...any) {
	l.console.Warn(msg, args...)
	l.file.Warn(msg, args...)
	l.gui.Warn(msg, args...)
}

func (l *Logger) DebugGui(msg string, args ...any) {
	l.console.Debug(msg, args...)
	l.file.Debug(msg, args...)
	l.gui.Debug(msg, args...)
}

// WithModule para agregar módulo
func (l *Logger) WithModule(module string) *Logger {
	return &Logger{
		console:     l.console.With(slog.String(LoggerAttributeKeyModule, module)),
		file:        l.file.With(slog.String(LoggerAttributeKeyModule, module)),
		gui:         l.gui.With(slog.String(LoggerAttributeKeyModule, module)),
		level:       l.level,
		fileCloser:  l.fileCloser,
		logFilePath: l.logFilePath,
	}
}

// WithFields para agregar campos adicionales
func (l *Logger) WithFields(fields ...any) *Logger {
	return &Logger{
		console:     l.console.With(fields...),
		file:        l.file.With(fields...),
		gui:         l.gui.With(fields...),
		level:       l.level,
		fileCloser:  l.fileCloser,
		logFilePath: l.logFilePath,
	}
}

// Propiedades públicas para compatibilidad si las necesitas
func (l *Logger) Console() *slog.Logger { return l.console }
func (l *Logger) File() *slog.Logger    { return l.file }
func (l *Logger) Gui() *slog.Logger     { return l.gui }
