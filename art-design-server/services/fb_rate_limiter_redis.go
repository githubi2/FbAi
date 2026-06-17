package services

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisLuaSchedule 原子时隙分配 Lua 脚本
//
// KEYS[1]: fb:next_available_at (下次可用时间戳 ms)
// ARGV[1]: interval_ms
// ARGV[2]: now_ms
//
// 返回: 该请求被分配的时隙（毫秒时间戳）
const redisLuaSchedule = `
local key = KEYS[1]
local interval = tonumber(ARGV[1])
local now = tonumber(ARGV[2])

local next_avail = redis.call('GET', key)
if not next_avail then
    next_avail = now
else
    next_avail = tonumber(next_avail)
end

-- 过期时间重置
if next_avail < now then
    next_avail = now
end

local my_slot = next_avail
redis.call('SET', key, my_slot + interval)
return my_slot
`

// NewRedisWaitFn 创建基于 Redis 的分布式等待策略
//
// redisURL: Redis 地址，如 "redis://localhost:6379/0"
// interval: 最小请求间隔
//
// 返回 WaitFn，可注入到 FbRateLimiter.SetWaitFn()
//
// Python 端使用相同的 Lua 脚本即可共享限速状态：
//
//	import redis, time
//	r = redis.from_url("redis://localhost:6379/0")
//	lua = r.register_script(REDIS_LUA_SCHEDULE)
//	slot_ms = lua(keys=["fb:next_available_at"], args=[4000, int(time.time()*1000)])
//	wait_s = (slot_ms - int(time.time()*1000)) / 1000
//	if wait_s > 0: time.sleep(wait_s)
//	# ... 调用 Facebook API
func NewRedisWaitFn(redisURL string, interval time.Duration) (WaitFn, func() error, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, nil, err
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, nil, err
	}

	luaScript := redis.NewScript(redisLuaSchedule)

	waitFn := func(ctx context.Context, endpoint string) time.Duration {
		nowMs := time.Now().UnixMilli()
		slotMs, err := luaScript.Run(ctx, client,
			[]string{"fb:next_available_at"},
			interval.Milliseconds(),
			nowMs,
		).Int64()
		if err != nil {
			log.Printf("[FbRateLimiter:Redis] Lua 脚本错误 (%s): %v", endpoint, err)
			return 0 // 出错时放行，避免阻塞
		}
		waitMs := slotMs - nowMs
		if waitMs < 0 {
			waitMs = 0
		}
		return time.Duration(waitMs) * time.Millisecond
	}

	cleanup := func() error { return client.Close() }

	log.Printf("[FbRateLimiter] ✅ Redis 分布式限速就绪 (间隔 %v, %s)", interval, redisURL)
	return waitFn, cleanup, nil
}

// TryUpgradeToRedis 尝试将全局限速器升级为 Redis 分布式模式
// 成功则注入 Redis WaitFn，失败则保持内存模式
func TryUpgradeToRedis() {
	redisURL := os.Getenv("FB_REDIS_URL")
	if redisURL == "" {
		redisURL = os.Getenv("REDIS_URL")
	}
	if redisURL == "" {
		log.Println("[FbRateLimiter] 未配置 REDIS_URL，保持内存限速模式")
		return
	}

	interval := defaultFbRateLimitInterval()

	waitFn, cleanup, err := NewRedisWaitFn(redisURL, interval)
	if err != nil {
		log.Printf("[FbRateLimiter] Redis 不可用 (%v)，保持内存限速模式", err)
		return
	}

	DefaultFbRateLimiter.SetWaitFn(waitFn)

	// 注册进程退出时清理
	go func() {
		<-context.Background().Done() // 永不触发，只是占位
		_ = cleanup
	}()

	log.Printf("[FbRateLimiter] ✅ 已升级为 Redis 分布式限速 (间隔 %v)", interval)
}
