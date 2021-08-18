package runner

import (
	"context"
	"log"
	"net/http"
	"runtime"
)

type HttpServer struct {
	config  *Config
	handler http.Handler
	httpSrv *http.Server
}

func NewHttpServer(handler http.Handler) *HttpServer {
	s := &HttpServer{
		handler: handler,
	}
	return s
}

// Run 监听端口
func (s *HttpServer) Run(opts ...Option) error {
	s.config = new(Config).Init().WithOptions(opts...)

	s.httpSrv = &http.Server{
		Addr:         s.config.ListenAddr,
		Handler:      s.handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}
	log.Printf("http server %s\n", s.config)
	s.httpSrv.RegisterOnShutdown(func() {
		log.Printf("current goroutine number: %d", runtime.NumGoroutine())
	})
	return s.httpSrv.ListenAndServe()
}

// Shutdown 优雅关闭
func (s *HttpServer) Shutdown(ctx context.Context) error {
	var err error
	if s.httpSrv != nil {
		err = s.httpSrv.Shutdown(ctx)
	}
	log.Printf("http server shutdown %v\n", err)
	return err
}
