package runner

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
)

type HttpServer struct {
	c       *Config
	handler http.Handler
	httpSrv *http.Server
}

func NewHttpServer(handler http.Handler) *HttpServer {
	srv := &HttpServer{
		handler: handler,
	}
	return srv
}

// Run
func (s *HttpServer) Run(c *Config) error {
	s.c = c
	s.httpSrv = &http.Server{
		Addr:         s.c.ListenAddr,
		Handler:      s.handler,
		ReadTimeout:  s.c.ReadTimeout,
		WriteTimeout: s.c.WriteTimeout,
	}
	log.Printf("http server %s\n", s.c)

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(fmt.Sprintf("http ListenAndServe %v", err))
			} else {
				log.Printf("http ListenAndServe %v\n", err)
			}
		}
	}()

	s.httpSrv.RegisterOnShutdown(func() {
		log.Printf("current goroutine number: %d", runtime.NumGoroutine())
	})

	return nil
}

// Shutdown
func (s *HttpServer) Shutdown(ctx context.Context) error {
	var err error
	if s.httpSrv != nil {
		err = s.httpSrv.Shutdown(ctx)
	}
	log.Printf("http server shutdown %v\n", err)
	return err
}
