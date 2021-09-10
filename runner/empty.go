package runner

import (
	"context"
	"fmt"
	"time"
)

type EmptyServer struct {
}

func (s *EmptyServer) Run(cfg *Config) error {
	go func() {
		execLoop := func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					fmt.Println("EmptyServer ctx.Done")
					return
				default:
					fmt.Println("empty running")
					time.Sleep(1 * time.Second)
				}
			}
		}
		execLoop(cfg.Context)
	}()

	return nil
}

func (s *EmptyServer) Shutdown(ctx context.Context) error {
	fmt.Println("EmptyServer Shutdown")
	return nil
}
