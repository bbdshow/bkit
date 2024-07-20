package bkit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

// RunServerAndListenExitSignal 运行服务，并监听退出信号
func RunServerAndListenExitSignal(run, shutdown, deallocate func(ctx context.Context) error) error {
	if run == nil || shutdown == nil {
		return fmt.Errorf("run | shutdown function required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errC := make(chan error, 1)
	go func() {
		errC <- ErrMulti(run(ctx))
	}()

	err := <-errC
	if err != nil {
		cancel()
		return err
	}

	// 监听退出信号
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	<-exitSignal
	log.Printf("receive exit signal\n")

	// 设置超时，防止Kill进程无法退出
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 关闭服务
	err = ErrMulti(shutdown(ctx))
	// 回收资源
	if deallocate != nil {
		// 设置超时
		err = ErrMulti(deallocate(ctx))
	}
	return err
}

type HTTPServerConf struct {
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (in *HTTPServerConf) Validate() error {
	if in.ListenAddr == "" {
		return fmt.Errorf("ListenAddr required")
	}
	if in.ReadTimeout <= 0 {
		in.ReadTimeout = 10 * time.Minute
	}
	if in.WriteTimeout <= 0 {
		in.WriteTimeout = 10 * time.Minute
	}
	return nil
}

func (c *HTTPServerConf) String() string {
	return fmt.Sprintf("HTTP监听地址: %s 服务读超时: %s 服务写超时: %s", c.ListenAddr, c.ReadTimeout, c.WriteTimeout)
}

type HTTPServer struct {
	cfg        *HTTPServerConf
	handler    http.Handler
	httpServer *http.Server
}

func NewHTTPServer(cfg *HTTPServerConf, handler http.Handler) *HTTPServer {
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	s := &HTTPServer{
		cfg:     cfg,
		handler: handler,
	}
	return s
}

func (s *HTTPServer) Run(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      s.handler,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}
	log.Printf("http server %s\n", s.cfg)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(fmt.Sprintf("http ListenAndServe %v", err))
			} else {
				log.Printf("http ListenAndServe %v\n", err)
			}
		}
	}()

	s.httpServer.RegisterOnShutdown(func() {
		log.Printf("current goroutine number: %d", runtime.NumGoroutine())
	})
	return nil
}

// Shutdown
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	var err error
	if s.httpServer != nil {
		err = s.httpServer.Shutdown(ctx)
	}
	log.Printf("http server shutdown %v\n", err)
	return err
}

type TaskServer struct {
	ctx    context.Context
	cancel func()

	wg *sync.WaitGroup

	afters []timeAfterFunc
	c      *cron.Cron
}

func NewTaskServer() *TaskServer {
	s := &TaskServer{
		wg:     &sync.WaitGroup{},
		afters: make([]timeAfterFunc, 0),
	}
	return s
}

type timeAfterFunc struct {
	d      time.Duration
	isOnce bool
	fn     func(ctx context.Context)
}

func (s *TaskServer) Run(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	for _, v := range s.afters {
		// register and run time.AfterFunc
		exec := func(ctx context.Context, wg *sync.WaitGroup, fn timeAfterFunc) {
			for {
				time.Sleep(fn.d)
				select {
				case <-ctx.Done():
					return
				default:
					func() {
						// WaitGroup is reused before previous Wait has returned
						wg.Add(1)
						fn.fn(ctx)
						wg.Done()
					}()
					if fn.isOnce {
						return
					}
				}
			}
		}
		go exec(s.ctx, s.wg, v)
	}

	if s.c != nil {
		s.c.Start()
	}

	return nil
}

func (s *TaskServer) Shutdown(ctx context.Context) error {
	s.cancel()
	if s.c != nil {
		<-s.c.Stop().Done()
	}
	// 等待所有任务结束
	s.wg.Wait()

	log.Printf("task server shutdown\n")
	return nil
}

// you can use AddTickerTimeAfterFunc replace this
// Deprecated
func (s *TaskServer) AddTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	s.afters = append(s.afters, timeAfterFunc{
		d:  d,
		fn: fn,
	})
	return nil
}

// AddOnceTimeAfterFunc after d, once exec
func (s *TaskServer) AddOnceTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	s.afters = append(s.afters, timeAfterFunc{
		d:      d,
		isOnce: true,
		fn:     fn,
	})
	return nil
}

// AddTickerTimeAfterFunc ticker d, loop exec
func (s *TaskServer) AddTickerTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	s.afters = append(s.afters, timeAfterFunc{
		d:      d,
		isOnce: false,
		fn:     fn,
	})
	return nil
}

func (s *TaskServer) AddCronFunc(spec string, fn func()) error {
	if s.c == nil {
		s.c = cron.New(cron.WithSeconds())
	}
	_, err := s.c.AddFunc(spec, fn)
	if err != nil {
		return err
	}
	return nil
}
