package logman

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/rehiy/pango/onquit"
)

type Logger struct {
	name   string
	logger *slog.Logger
	ctx    context.Context
}

func Named(name string) *Logger {
	return NamedWithContext(name, context.Background())
}

func NamedWithContext(name string, ctx context.Context) *Logger {
	return &Logger{
		name:   name,
		logger: NewLogger(name),
		ctx:    ctx,
	}
}

func (l *Logger) log(level slog.Level, msg string, args ...any) {
	args = append([]any{"Logger", l.name}, args...)
	l.logger.Log(l.ctx, level, msg, args...)
}

// WithContext sets the context

func (l *Logger) WithContext(ctx context.Context) *Logger {
	l.ctx = ctx
	return l
}

// Debug logs a debug message

func (l *Logger) Debug(msg string, args ...any) {
	l.log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Debugf(msg string, args ...any) {
	l.log(slog.LevelDebug, fmt.Sprintf(msg, args...))
}

// Info logs an info message

func (l *Logger) Info(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Infof(msg string, args ...any) {
	l.log(slog.LevelInfo, fmt.Sprintf(msg, args...))
}

// Warn logs a warning message

func (l *Logger) Warn(msg string, args ...any) {
	l.log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Warnf(msg string, args ...any) {
	l.log(slog.LevelWarn, fmt.Sprintf(msg, args...))
}

// Error logs an error message

func (l *Logger) Error(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

func (l *Logger) Errorf(msg string, args ...any) {
	l.log(slog.LevelError, fmt.Sprintf(msg, args...))
}

// Fatal logs a fatal message

func (l *Logger) Fatal(msg string, args ...any) {
	onquit.CallQuitFuncs() // 调用所有退出函数
	l.log(slog.LevelError, msg, args...)
	os.Exit(1)
}

func (l *Logger) Fatalf(msg string, args ...any) {
	onquit.CallQuitFuncs() // 调用所有退出函数
	l.log(slog.LevelError, fmt.Sprintf(msg, args...))
	os.Exit(1)
}
