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
	srv := &TaskServer{
		wg:     &sync.WaitGroup{},
		afters: make([]timeAfterFunc, 0),
	}
	return srv
}

type timeAfterFunc struct {
	d  time.Duration
	fn func(ctx context.Context)
}

func (srv *TaskServer) Run(opts ...Option) error {
	srv.config = new(Config).Init().WithOptions(opts...)
	srv.config.Context, srv.cancel = context.WithCancel(srv.config.Context)

	for _, v := range srv.afters {
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
				}
			}
		}
		go exec(srv.config.Context, srv.wg, v)
	}

	if srv.c != nil {
		srv.c.Start()
	}

	return nil
}

func (srv *TaskServer) Shutdown(ctx context.Context) error {
	srv.cancel()
	// waiting exec over
	if srv.c != nil {
		<-srv.c.Stop().Done()
	}
	srv.wg.Wait()

	log.Printf("task server shutdown\n")
	return nil
}

func (srv *TaskServer) AddTimeAfterFunc(d time.Duration, fn func(ctx context.Context)) error {
	if d <= 0 {
		return fmt.Errorf("d required")
	}
	if fn == nil {
		return fmt.Errorf("func required")
	}
	srv.afters = append(srv.afters, timeAfterFunc{
		d:  d,
		fn: fn,
	})
	return nil
}

func (srv *TaskServer) AddCronFunc(spec string, fn func()) error {
	if srv.c == nil {
		srv.c = cron.New(cron.WithSeconds())
	}
	_, err := srv.c.AddFunc(spec, fn)
	if err != nil {
		return err
	}
	return nil
}
