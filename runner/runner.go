package runner

import (
	"context"
	"fmt"
	"github.com/bbdshow/bkit/errc"
	"github.com/bbdshow/bkit/gen/defval"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Run(...Option) error
	Shutdown(context.Context) error
}

// Run 运行服务进程，一般放到main.go 最后一行执行
func Run(server Server, deallocFunc func() error, opts ...Option) error {
	if server == nil {
		panic("server required")
	}
	var err error

	go func() {
		if err := server.Run(opts...); err != nil {
			err = errc.MultiError(err)
		}
	}()

	// 拦截退出信号
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	<-exitSignal

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// server shutdown
	err = errc.MultiError(server.Shutdown(ctx))

	// 释放资源
	if deallocFunc != nil {
		err = errc.MultiError(deallocFunc())
	}
	return err
}

type Option interface {
	apply(*Config)
}

type optionFunc func(*Config)

func (opt optionFunc) apply(config *Config) {
	opt(config)
}

type Config struct {
	ListenAddr   string        `defval:"0.0.0.0:8080"` // host:port
	ReadTimeout  time.Duration `defval:"10m"`
	WriteTimeout time.Duration `defval:"10m"`
	Context      context.Context
}

func (c *Config) init() *Config {
	if err := defval.ParseDefaultVal(c); err != nil {
		panic(err)
	}
	if c.Context == nil {
		c.Context = context.Background()
	}
	return c
}

func (c *Config) String() string {
	return fmt.Sprintf("listenAddr: %s readTimeout: %s writeTimeout: %s", c.ListenAddr, c.ReadTimeout, c.WriteTimeout)
}

func (c *Config) withOptions(opts ...Option) *Config {
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func WithListenAddr(addr string) Option {
	return optionFunc(func(config *Config) {
		config.ListenAddr = addr
	})
}

func WithReadTimeout(timeout time.Duration) Option {
	return optionFunc(func(config *Config) {
		if timeout > 0 {
			config.ReadTimeout = timeout
		}
	})
}

func WithWriteTimeout(timeout time.Duration) Option {
	return optionFunc(func(config *Config) {
		if timeout > 0 {
			config.WriteTimeout = timeout
		}
	})
}

func WithContext(ctx context.Context) Option {
	return optionFunc(func(config *Config) {
		if ctx != nil {
			config.Context = ctx
		}
	})
}
