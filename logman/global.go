package logman

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rehiy/pango/onquit"
)

var (
	Debug = slog.Debug
	Info  = slog.Info
	Warn  = slog.Warn
	Error = slog.Error
)

func Debugf(msg string, args ...any) {
	Debug(fmt.Sprintf(msg, args...))
}

func Infof(msg string, args ...any) {
	Info(fmt.Sprintf(msg, args...))
}

func Warnf(msg string, args ...any) {
	Warn(fmt.Sprintf(msg, args...))
}

func Errorf(msg string, args ...any) {
	Error(fmt.Sprintf(msg, args...))
}

func Fatal(msg string, args ...any) {
	onquit.CallQuitFuncs()
	Error(msg, args...)
	os.Exit(1)
}

func Fatalf(msg string, args ...any) {
	Fatal(fmt.Sprintf(msg, args...))
}
