package main

import (
	"context"
	"fmt"
	"github.com/bbdshow/bkit/ginutil"
	"github.com/bbdshow/bkit/runner"
	"os"
	"time"
)

func main() {
	ctx := context.Background()

	go func() {
		err := runner.Run(runnerTask(), func() error {
			return nil
		}, runner.WithContext(ctx))
		if err != nil {
			fmt.Println("exit", err)
		}
	}()

	go func() {
		if err := runner.RunServer(runner.NewGrpcServer(),
			runner.WithContext(ctx),
			runner.WithListenAddr("0.0.0.0:18081")); err != nil {
			os.Exit(1)
		}
	}()

	go func() {
		midFlag := ginutil.MStd | ginutil.MPprof
		httpHandler := ginutil.DefaultEngine(midFlag)
		if err := runner.RunServer(runner.NewHttpServer(httpHandler),
			runner.WithContext(ctx),
			runner.WithListenAddr("0.0.0.0:18080")); err != nil {
			os.Exit(1)
		}
		//if err := runner.RunServer(runner.NewHttpServer(http.NotFoundHandler()),
		//	runner.WithContext(ctx),
		//	runner.WithListenAddr("0.0.0.0:18080")); err != nil {
		//	os.Exit(1)
		//}
	}()

	time.Sleep(1 * time.Second)

	if err := runner.RunServer(new(runner.EmptyServer), runner.WithContext(ctx)); err != nil {
		os.Exit(1)
	}

	time.Sleep(2 * time.Second)
}

func runnerTask() runner.Server {
	srv := runner.NewTaskServer()
	if err := srv.AddTimeAfterFunc(time.Second, func(ctx context.Context) {
		execWithCtx(ctx)
	}); err != nil {
		panic(err)
	}
	return srv
}

func execWithCtx(ctx context.Context) {
	timer := time.NewTimer(5 * time.Second)
	select {
	case <-timer.C:
		fmt.Println("timer")
		timer.Stop()
	case <-ctx.Done():
		fmt.Println("ctx done")
	}
}
