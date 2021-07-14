package runner

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewTaskServer(t *testing.T) {
	ctx := context.Background()
	err := Run(runnerTask(ctx), func() error {
		return nil
	}, WithContext(ctx))
	if err != nil {
		fmt.Println("exit", err)
	}
}

func runnerTask(ctx context.Context) Server {
	srv := NewTaskServer()
	if err := srv.AddTimeAfterFunc(time.Second, func() {
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
