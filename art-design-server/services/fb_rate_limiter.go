package services

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"
)

// fbRateLimitRequest 限速队列中的请求
type fbRateLimitRequest struct {
	endpoint string
	fn       func() (interface{}, error)
	resultCh chan fbRateLimitResult
}

type fbRateLimitResult struct {
	data interface{}
	err  error
}

// WaitFn 外部等待策略（如 Redis 分布式限速）
// endpoint: 请求端点（用于日志）
// 返回: 需要等待的时长，0 表示无需等待
type WaitFn func(ctx context.Context, endpoint string) time.Duration

// FbRateLimiter Facebook API 请求限速器
// 使用单 goroutine + channel 实现串行队列
// 默认使用本地内存计时，可选注入 WaitFn 实现分布式限速（如 Redis）
type FbRateLimiter struct {
	minInterval time.Duration
	requests    chan *fbRateLimitRequest
	waitFn      WaitFn // nil = 使用本地计时
}

// DefaultFbRateLimiter 全局 FB 限速器实例（默认内存模式，间隔 4 秒）
var DefaultFbRateLimiter = newFbRateLimiter(defaultFbRateLimitInterval(), nil)

func defaultFbRateLimitInterval() time.Duration {
	if v := os.Getenv("FB_RATE_LIMIT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return 4 * time.Second
}

func newFbRateLimiter(minInterval time.Duration, waitFn WaitFn) *FbRateLimiter {
	rl := &FbRateLimiter{
		minInterval: minInterval,
		requests:    make(chan *fbRateLimitRequest, 128),
		waitFn:      waitFn,
	}
	go rl.worker()
	return rl
}

// SetWaitFn 设置外部等待策略（如 Redis 分布式限速）
// 必须在任何请求发出之前调用
func (rl *FbRateLimiter) SetWaitFn(fn WaitFn) {
	rl.waitFn = fn
}

// Do 将请求加入限速队列，阻塞等待执行完毕后返回结果
func (rl *FbRateLimiter) Do(
	ctx context.Context,
	endpoint string,
	fn func() (interface{}, error),
) (interface{}, error) {
	resultCh := make(chan fbRateLimitResult, 1)
	req := &fbRateLimitRequest{
		endpoint: endpoint,
		fn:       fn,
		resultCh: resultCh,
	}

	select {
	case rl.requests <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-resultCh:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// worker 串行处理队列中的请求
func (rl *FbRateLimiter) worker() {
	var lastRequestAt time.Time

	for req := range rl.requests {
		// 如果有外部等待策略（如 Redis），使用它
		if rl.waitFn != nil {
			waitDuration := rl.waitFn(context.Background(), req.endpoint)
			if waitDuration > 0 {
				log.Printf("[FbRateLimiter] \"%s\" 排队等待 %v", req.endpoint, waitDuration.Round(time.Millisecond))
				time.Sleep(waitDuration)
			}
		} else {
			// 默认本地计时
			elapsed := time.Since(lastRequestAt)
			if elapsed < rl.minInterval {
				waitTime := rl.minInterval - elapsed
				log.Printf("[FbRateLimiter] \"%s\" 排队等待 %v", req.endpoint, waitTime.Round(time.Millisecond))
				time.Sleep(waitTime)
			}
		}

		lastRequestAt = time.Now()
		data, err := req.fn()
		req.resultCh <- fbRateLimitResult{data: data, err: err}
	}
}
