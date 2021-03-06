package runner

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type GrpcServer struct {
	c      *Config
	server *grpc.Server

	runAfters []func(s *grpc.Server) error
}

func NewGrpcServer() *GrpcServer {
	s := &GrpcServer{
		runAfters: make([]func(s *grpc.Server) error, 0),
	}
	return s
}

func (s *GrpcServer) Run(c *Config) error {
	s.c = c
	listen, err := net.Listen("tcp", s.c.ListenAddr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	s.server = server

	log.Printf("grpc server %s\n", s.c)
	for _, fn := range s.runAfters {
		if fn != nil {
			if err := fn(s.server); err != nil {
				return err
			}
		}
	}

	go func() {
		if err := server.Serve(listen); err != nil {
			panic(fmt.Sprintf("grpc serve listen %v", err))
		}
	}()

	return nil
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		s.server.Stop()
	}
	log.Printf("grpc server shutdown \n")
	return nil
}

func (s *GrpcServer) RunAfter(fn func(s *grpc.Server) error) {
	if fn != nil {
		s.runAfters = append(s.runAfters, fn)
	}
}

func (s *GrpcServer) Server() *grpc.Server {
	return s.server
}
