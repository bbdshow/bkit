package main

import (
	"context"
	"fmt"
	"github.com/bbdshow/bkit/runner"
	"time"
)

func main() {
	ctx := context.Background()
	err := runner.Run(runnerTask(), func() error {
		return nil
	}, runner.WithContext(ctx))
	if err != nil {
		fmt.Println("exit", err)
	}
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
