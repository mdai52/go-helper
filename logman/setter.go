package logman

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

var config = &Config{}

type Config struct {
	Level    string `note:"日志级别 debug|info|warn|error"`
	Target   string `note:"日志输出设备 both|file|null|stdout|stderr"`
	Storage  string `note:"日志文件存储目录"`
	Filename string `note:"默认日志文件名"`
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

func NewLogger(name string) *slog.Logger {
	var level slog.Level
	level.UnmarshalText([]byte(config.Level))

	option := &slog.HandlerOptions{
		Level: level,
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
		f = filepath.Join(config.Storage, name) + ".log"
	}

	if d := filepath.Dir(f); d != "" && d != "." {
		os.MkdirAll(d, 0755)
	}

	return &lumberjack.Logger{
		Compress:   true, // 是否压缩/归档旧文件
		Filename:   f,    // 日志文件位置
		MaxSize:    100,  // 单个日志文件最大值(单位：MB)
		MaxBackups: 21,   // 保留旧文件的最大个数
		MaxAge:     7,    // 保留旧文件的最大天数
	}
}
