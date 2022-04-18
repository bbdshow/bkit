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
	s := NewTaskServer()
	runCount := 0
	if err := s.AddTimeAfterFunc(time.Second, func(ctx context.Context) {
		runCount++
		//execTickerWithCtx(ctx, "AddTimeAfterFunc")
		fmt.Println(runCount)
	}); err != nil {
		panic(err)
	}

	runCount2 := 0
	if err := s.AddOnceTimeAfterFunc(time.Second, func(ctx context.Context) {
		runCount2++
		execTickerWithCtx(ctx, "AddOnceTimeAfterFunc")
		fmt.Println(runCount2)
	}); err != nil {
		panic(err)
	}

	return s
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

func execTickerWithCtx(ctx context.Context, method string) {
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			select {
			case <-ctx.Done():
				fmt.Println("ctx done")
			default:
				fmt.Println("run ticker ", method)
			}
		}
	}()
}
