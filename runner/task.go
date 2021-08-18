package runner

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"sync"
	"time"
)

type TaskServer struct {
	config *Config

	cancel func()
	wg     *sync.WaitGroup

	afters []timeAfterFunc
	c      *cron.Cron
}

func NewTaskServer() *TaskServer {
	s := &TaskServer{
		wg:     &sync.WaitGroup{},
		afters: make([]timeAfterFunc, 0),
	}
	return s
}

type timeAfterFunc struct {
	d  time.Duration
	fn func(ctx context.Context)
}

func (s *TaskServer) Run(opts ...Option) error {
	s.config = new(Config).Init().WithOptions(opts...)
	s.config.Context, s.cancel = context.WithCancel(s.config.Context)

	for _, v := range s.afters {
		// 注册并运行 time.AfterFunc 任务
		exec := func(ctx context.Context, wg *sync.WaitGroup, fn timeAfterFunc) {
			for {
				time.Sleep(fn.d)
				select {
				case <-ctx.Done():
					return
				default:
					func() {
						// WaitGroup is reused before previous Wait has returned
						wg.Add(1)
						fn.fn(ctx)
						wg.Done()
					}()
				}
			}
		}
		// 运行
		go exec(s.config.Context, s.wg, v)
	}

	if s.c != nil {
		s.c.Start()
	}

	return nil
}

func (s *TaskServer) Shutdown(ctx context.Context) error {
	s.cancel()
	// 等待任务执行完
	if s.c != nil {
		<-s.c.Stop().Done()
	}
	s.wg.Wait()

	log.Printf("task server shutdown\n")
	return nil
}

func (s *TaskServer) AddTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	s.afters = append(s.afters, timeAfterFunc{
		d:  d,
		fn: fn,
	})
	return nil
}

func (s *TaskServer) AddCronFunc(spec string, fn func()) error {
	if s.c == nil {
		s.c = cron.New(cron.WithSeconds())
	}
	_, err := s.c.AddFunc(spec, fn)
	if err != nil {
		return err
	}
	return nil
}
