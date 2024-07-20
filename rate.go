package bkit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

var Limiter *LimiterUtil

func init() {
	Limiter = NewLimiterUtil()
}

// LimiterUtil 限流器
type LimiterUtil struct {
	mutex sync.RWMutex

	limiterMap map[string]*rate.Limiter
}

func NewLimiterUtil() *LimiterUtil {
	return &LimiterUtil{
		limiterMap: make(map[string]*rate.Limiter),
	}
}

// Register 注册限流器, r 为每秒请求数, b 为桶的容量(短时间内允许的最大请求数)
func (lim *LimiterUtil) Register(key string, r, b int) {
	lim.mutex.Lock()
	defer lim.mutex.Unlock()

	limiter, ok := lim.limiterMap[key]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(r), b)
		lim.limiterMap[key] = limiter
	}
	if limiter.Limit() != rate.Limit(r) {
		limiter.SetLimit(rate.Limit(r))
	}
	if limiter.Burst() != b {
		limiter.SetBurst(b)
	}
}

// Allow 是否允许通过
func (lim *LimiterUtil) Allow(key string) bool {
	lim.mutex.RLock()
	defer lim.mutex.RUnlock()

	limiter, ok := lim.limiterMap[key]
	if !ok {
		return true
	}
	return limiter.Allow()
}

// Wait 等待直到有请求可以通过
func (lim *LimiterUtil) Wait(ctx context.Context, key string) error {
	lim.mutex.RLock()
	defer lim.mutex.RUnlock()

	limiter, ok := lim.limiterMap[key]
	if !ok {
		return nil
	}
	return limiter.Wait(ctx)
}
