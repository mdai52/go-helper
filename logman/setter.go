package logman

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var config = &Config{}

type Config struct {
	Level    string `note:"日志级别 debug|info|warn|error" json:"level"`
	Target   string `note:"日志输出设备 both|file|null|stdout|stderr" json:"target"`
	Storage  string `note:"日志文件存储目录" json:"storage"`
	Filename string `note:"默认日志文件名" json:"filename"`
}

func Default() *slog.Logger {
	return slog.Default()
}

func SetDefault(args *Config) {
	config.Level = args.Level
	config.Target = args.Target
	config.Storage = args.Storage

	slog.SetDefault(NewLogger(args.Filename))
}

// replaceAttr 日志属性替换函数
func replaceAttr(a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		a.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05"))
	case "Logger":
		// 清理 Logger 字段值开头的 . 和 /
		a.Value = slog.StringValue(strings.TrimLeft(a.Value.String(), "./"))
	}
	return a
}

func NewLogger(name string) *slog.Logger {
	var level slog.Level
	level.UnmarshalText([]byte(config.Level))

	option := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return replaceAttr(a)
		},
	}

	handler := slog.NewTextHandler(AutoWriter(name), option)
	return slog.New(handler)
}

func AutoWriter(name string) io.Writer {
	switch config.Target {
	case "file":
		return FileWriter(name)
	case "both":
		return io.MultiWriter(os.Stdout, FileWriter(name))
	case "null":
		return io.Discard
	case "stderr":
		return os.Stderr
	default:
		return os.Stdout
	}
}

func FileWriter(name string) *lumberjack.Logger {
	f := name + ".log"
	if !strings.HasPrefix(name, "/") {
		f = filepath.Join(config.Storage, f)
	}

	if d := filepath.Dir(f); d != "" && d != "." {
		os.MkdirAll(d, 0755)
	}

	return &lumberjack.Logger{
		Compress:   true,
		Filename:   f,
		MaxSize:    100,
		MaxBackups: 21,
		MaxAge:     7,
	}
}
