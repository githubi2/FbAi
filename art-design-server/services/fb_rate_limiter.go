package services

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WaitFn 外部等待策略（如 Redis 分布式限速）
// endpoint: 请求端点（用于日志）
// 返回: 需要等待的时长，0 表示无需等待
type WaitFn func(ctx context.Context, endpoint string) time.Duration

// FbRateLimiter Facebook API 请求限速器（多账户并发安全）
// 设计（依据官方 BUC 限速文档：限额按应用/广告账户维度，跨账户调用相互独立）：
//   - 按广告账户（act_xxx）独立限速：同一账户串行且间隔 keyInterval，不同账户并行
//   - 全局并发上限 concurrency：防止瞬时请求过多触发平台级节流
//
// 默认 keyInterval=1s/账户，并发=10（可通过 FB_RATE_KEY_MS / FB_FETCH_CONCURRENCY 覆盖）
type FbRateLimiter struct {
	keyInterval time.Duration
	concurrency int
	waitFn      WaitFn // nil = 使用本地计时

	mu         sync.Mutex
	keyTimers  map[string]time.Time // key → 上次请求时间
	lastGlobal time.Time            // 全局兜底计时（waitFn 为空且无 key 时）
	sem        chan struct{}        // 全局并发信号量
}

// DefaultFbRateLimiter 全局 FB 限速器实例
var DefaultFbRateLimiter = newFbRateLimiter(defaultFbRateKeyInterval(), defaultFbFetchConcurrency(), nil)

func defaultFbRateKeyInterval() time.Duration {
	if v := os.Getenv("FB_RATE_KEY_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return 1 * time.Second
}

func defaultFbFetchConcurrency() int {
	if v := os.Getenv("FB_FETCH_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			return n
		}
	}
	return 10
}

func newFbRateLimiter(keyInterval time.Duration, concurrency int, waitFn WaitFn) *FbRateLimiter {
	return &FbRateLimiter{
		keyInterval: keyInterval,
		concurrency: concurrency,
		waitFn:      waitFn,
		keyTimers:   make(map[string]time.Time),
		sem:         make(chan struct{}, concurrency),
	}
}

// SetWaitFn 设置外部等待策略（如 Redis 分布式限速）
func (rl *FbRateLimiter) SetWaitFn(fn WaitFn) {
	rl.waitFn = fn
}

// rateLimitKey 从 endpoint 提取限速 key（广告账户 act_xxx；无则用 global）
// endpoint 形如: /v26.0/act_2947189472297239/campaigns
func rateLimitKey(endpoint string) string {
	parts := strings.Split(strings.TrimPrefix(endpoint, "/"), "/")
	if len(parts) >= 3 && strings.HasPrefix(parts[1], "act_") {
		return parts[1]
	}
	return "global"
}

// Do 将请求加入限速，阻塞等待后执行 fn（同账户串行限速 + 全局并发上限）
func (rl *FbRateLimiter) Do(
	ctx context.Context,
	endpoint string,
	fn func() (interface{}, error),
) (interface{}, error) {
	key := rateLimitKey(endpoint)

	// 1. 全局并发槽
	select {
	case rl.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-rl.sem }()

	// 2. 限速等待（外部策略优先，否则按 key 间隔）
	var wait time.Duration
	rl.mu.Lock()
	last, ok := rl.keyTimers[key]
	rl.mu.Unlock()
	if rl.waitFn != nil {
		wait = rl.waitFn(ctx, endpoint)
	} else {
		if ok {
			remain := rl.keyInterval - time.Since(last)
			if remain > 0 {
				wait = remain
			}
		}
	}
	if wait > 0 {
		log.Printf("[FbRateLimiter] \"%s\" 排队等待 %v (key=%s)", endpoint, wait.Round(time.Millisecond), key)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 3. 记录时间并执行
	rl.mu.Lock()
	rl.keyTimers[key] = time.Now()
	rl.lastGlobal = time.Now()
	rl.mu.Unlock()

	return fn()
}
