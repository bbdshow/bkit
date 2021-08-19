package runner

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type GrpcServer struct {
	c      *Config
	server *grpc.Server
}

func NewGrpcServer() *GrpcServer {
	s := &GrpcServer{}
	return s
}

func (s *GrpcServer) Run(opts ...Option) error {
	s.c = new(Config).Init().WithOptions(opts...)
	listen, err := net.Listen("tcp", s.c.ListenAddr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	s.server = server

	log.Printf("grpc server %s\n", s.c)
	if err := server.Serve(listen); err != nil {
		return err
	}
	return nil
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		s.server.Stop()
	}
	log.Printf("grpc server shutdown \n")
	return nil
}
