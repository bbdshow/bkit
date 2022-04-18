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
	d      time.Duration
	isOnce bool
	fn     func(ctx context.Context)
}

func (s *TaskServer) Run(c *Config) error {
	s.config = c
	s.config.Context, s.cancel = context.WithCancel(s.config.Context)

	for _, v := range s.afters {
		// register and run time.AfterFunc
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
					if fn.isOnce {
						return
					}
				}
			}
		}
		go exec(s.config.Context, s.wg, v)
	}

	if s.c != nil {
		s.c.Start()
	}

	return nil
}

func (s *TaskServer) Shutdown(ctx context.Context) error {
	s.cancel()
	// waiting exec over
	if s.c != nil {
		<-s.c.Stop().Done()
	}
	s.wg.Wait()

	log.Printf("task server shutdown\n")
	return nil
}

//you can use AddTickerTimeAfterFunc replace this
//Deprecated
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

// AddOnceTimeAfterFunc after d, once exec
func (s *TaskServer) AddOnceTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	s.afters = append(s.afters, timeAfterFunc{
		d:      d,
		isOnce: true,
		fn:     fn,
	})
	return nil
}

// AddTickerTimeAfterFunc ticker d, loop exec
func (s *TaskServer) AddTickerTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	s.afters = append(s.afters, timeAfterFunc{
		d:      d,
		isOnce: false,
		fn:     fn,
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
