package bkit

import "time"

var Retry = RetryUtil{}

type RetryUtil struct{}

// RetryN 重试 n 次, sleep 为每次重试的间隔, n = 1 时不重试
func (RetryUtil) RetryN(n int, sleep time.Duration, fn func() error) error {
	if n <= 0 {
		n = 1
	}
	var err error
	for i := 0; i < n; i++ {
		err = fn()
		if err == nil {
			break
		}
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
	return err
}
