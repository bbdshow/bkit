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
	srv := &HttpServer{
		handler: handler,
	}
	return srv
}

// Run
func (srv *HttpServer) Run(opts ...Option) error {
	srv.config = new(Config).Init().WithOptions(opts...)

	srv.httpSrv = &http.Server{
		Addr:         srv.config.ListenAddr,
		Handler:      srv.handler,
		ReadTimeout:  srv.config.ReadTimeout,
		WriteTimeout: srv.config.WriteTimeout,
	}
	log.Printf("http server %s\n", srv.config)
	srv.httpSrv.RegisterOnShutdown(func() {
		log.Printf("current goroutine number: %d", runtime.NumGoroutine())
	})
	return srv.httpSrv.ListenAndServe()
}

// Shutdown
func (srv *HttpServer) Shutdown(ctx context.Context) error {
	var err error
	if srv.httpSrv != nil {
		err = srv.httpSrv.Shutdown(ctx)
	}
	log.Printf("http server shutdown %v\n", err)
	return err
}
