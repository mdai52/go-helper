package httpd

import (
	"context"
	"net/http"
	"time"

	"github.com/rehiy/pango/logman"
	"github.com/rehiy/pango/onquit"
)

type Config struct {
	Debug        bool
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Option func(*Config)

func WithDebug(debug bool) Option {
	return func(cfg *Config) {
		cfg.Debug = debug
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.ReadTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.WriteTimeout = d
	}
}

func Server(addr string, options ...Option) {
	cfg := &Config{
		Debug:        false,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	for _, opt := range options {
		opt(cfg)
	}

	Engine(cfg.Debug)
	server := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	onquit.Register(func() {
		// 创建一个剩余15秒超时的上下文
		logman.Warn("httpd will close within 9 seconds")
		ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()

		// 优雅地关闭服务而不中断任何活动连接
		if err := server.Shutdown(ctx); err != nil {
			logman.Warn("httpd forced to close", "error", err)
			server.Close()
		}
	})

	logman.Info("httpd start", "address", addr)
	if err := server.ListenAndServe(); err != nil {
		logman.Warn(err.Error())
	}
}
