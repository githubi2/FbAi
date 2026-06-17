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

// FbRateLimiter Facebook API 请求限速器
// 使用单 goroutine + channel 实现串行队列，
// 同一时刻只有一个请求在执行，请求之间至少间隔 minInterval
type FbRateLimiter struct {
	minInterval time.Duration
	requests    chan *fbRateLimitRequest
}

// DefaultFbRateLimiter 全局 FB 限速器实例（默认间隔 4 秒）
var DefaultFbRateLimiter = newFbRateLimiter(defaultFbRateLimitInterval())

func defaultFbRateLimitInterval() time.Duration {
	if v := os.Getenv("FB_RATE_LIMIT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return 4 * time.Second
}

func newFbRateLimiter(minInterval time.Duration) *FbRateLimiter {
	rl := &FbRateLimiter{
		minInterval: minInterval,
		requests:    make(chan *fbRateLimitRequest, 128),
	}
	go rl.worker()
	return rl
}

// Do 将请求加入限速队列，阻塞等待执行完毕后返回结果
// ctx 可用于取消排队中的请求（不会取消已发出的请求）
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

	// 发送到队列（支持 context 取消）
	select {
	case rl.requests <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 等待执行结果（支持 context 取消）
	select {
	case res := <-resultCh:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// worker 串行处理队列中的请求，确保请求间隔 >= minInterval
func (rl *FbRateLimiter) worker() {
	var lastRequestAt time.Time

	for req := range rl.requests {
		elapsed := time.Since(lastRequestAt)
		if elapsed < rl.minInterval {
			waitTime := rl.minInterval - elapsed
			log.Printf("[FbRateLimiter] \"%s\" 排队等待 %v", req.endpoint, waitTime.Round(time.Millisecond))
			time.Sleep(waitTime)
		}

		lastRequestAt = time.Now()
		data, err := req.fn()
		req.resultCh <- fbRateLimitResult{data: data, err: err}
	}
}
