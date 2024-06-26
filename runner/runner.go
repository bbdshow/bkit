package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbdshow/bkit/errc"
)

type Server interface {
	Run(*Config) error
	Shutdown(context.Context) error
}

func RunServer(server Server, opts ...Option) error {
	return Run(server, func() error { return nil }, opts...)
}

// Run
func Run(server Server, deallocFunc func() error, opts ...Option) error {
	if server == nil {
		panic("server required")
	}

	cfg := new(Config).Init().WithOptions(opts...)
	ctx, cancel := context.WithCancel(cfg.Context)
	defer cancel()

	cfg.Context = ctx

	errC := make(chan error, 1)
	go func() {
		errC <- errc.MultiError(server.Run(cfg))
	}()

	err := <-errC
	if err != nil {
		cancel()
		return err
	}

	// handler exit signal
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	<-exitSignal
	// server shutdown
	err = errc.MultiError(server.Shutdown(ctx))

	// deallocFunc
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
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Context      context.Context
}

func (c *Config) Init() *Config {
	if c.ListenAddr == "" {
		c.ListenAddr = ":8080"
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 10 * time.Minute
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 10 * time.Minute
	}

	if c.Context == nil {
		c.Context = context.Background()
	}
	return c
}

func (c *Config) String() string {
	return fmt.Sprintf("listenAddr: %s readTimeout: %s writeTimeout: %s", c.ListenAddr, c.ReadTimeout, c.WriteTimeout)
}

func (c *Config) WithOptions(opts ...Option) *Config {
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
