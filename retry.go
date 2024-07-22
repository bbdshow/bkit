package bkit

import "time"

var Retry = RetryUtil{}

type RetryUtil struct{}

// RetryN 重试 n 次, 至少执行一次, sleep 为每次重试的间隔, n = 1, 重试1次， 共执行2次
func (RetryUtil) RetryN(n uint, sleep time.Duration, fn func() error) error {
	var err error
	// i<=n 至少执行一次
	for i := uint(0); i <= n; i++ {
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
