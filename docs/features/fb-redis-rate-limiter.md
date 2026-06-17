# Python AI 自动化 — FB 限速集成指南

**日期**: 2026-06-18

## 概述

Go 后端和 Python AI 自动化共享同一个 Redis key `fb:next_available_at`，
通过相同的 Lua 原子脚本实现跨进程的 Facebook API 请求限速。

## 架构

```
Redis (fb:next_available_at)
    │
    ├── Go 后端 (fb_rate_limiter.go)
    │   └── 配置 REDIS_URL 后自动切换
    │
    └── Python AI (任意进程)
        └── 调用 fb_rate_limit() 后再请求 Facebook API
```

## Python 使用方式

### 安装依赖

```bash
pip install redis
```

### 代码

```python
import redis
import time
from contextlib import contextmanager

# ── Lua 脚本（与 Go 端完全一致） ──
REDIS_LUA_SCHEDULE = """
local key = KEYS[1]
local interval = tonumber(ARGV[1])
local now = tonumber(ARGV[2])

local next_avail = redis.call('GET', key)
if not next_avail then
    next_avail = now
else
    next_avail = tonumber(next_avail)
end

if next_avail < now then
    next_avail = now
end

local my_slot = next_avail
redis.call('SET', key, my_slot + interval)
return my_slot
"""


class FbRateLimiter:
    """Facebook API 分布式限速器

    用法:
        limiter = FbRateLimiter("redis://localhost:6379/0", interval_ms=4000)

        with limiter:
            # 调用 Facebook Graph API
            resp = requests.get("https://graph.facebook.com/v22.0/...")
    """

    def __init__(self, redis_url: str, interval_ms: int = 4000):
        self.client = redis.from_url(redis_url)
        self.interval_ms = interval_ms
        self.lua = self.client.register_script(REDIS_LUA_SCHEDULE)
        self._wait_sec = 0

    def acquire(self, endpoint: str = "") -> float:
        """获取时隙，返回等待的秒数（已自动 sleep）"""
        now_ms = int(time.time() * 1000)
        slot_ms = self.lua(
            keys=["fb:next_available_at"],
            args=[self.interval_ms, now_ms],
        )
        wait_ms = slot_ms - now_ms
        if wait_ms > 0:
            wait_sec = wait_ms / 1000
            if endpoint:
                print(f"[FbRateLimiter] '{endpoint}' 排队等待 {wait_sec:.1f}s")
            time.sleep(wait_sec)
            return wait_sec
        return 0

    def __enter__(self):
        self._wait_sec = self.acquire()
        return self

    def __exit__(self, *args):
        pass


# ═══════════════════════════════════════════════
# 使用示例
# ═══════════════════════════════════════════════

if __name__ == "__main__":
    limiter = FbRateLimiter("redis://localhost:6379/0", interval_ms=4000)

    # 方式 1: with 语句
    with limiter:
        # 调用 Facebook API
        # resp = requests.get("https://graph.facebook.com/v22.0/me/adaccounts",
        #                      params={"access_token": token, "fields": "name"})
        print("FB API call 1")

    # 方式 2: 显式调用
    limiter.acquire(endpoint="/me/adaccounts")
    # resp = requests.get(...)
    print("FB API call 2")
```

## 与 Go 后端的协同

| 配置 | Go | Python |
|------|:--:|:--:|
| Redis key | `fb:next_available_at` | `fb:next_available_at` |
| Lua 脚本 | `redisLuaSchedule` (Go 常量) | `REDIS_LUA_SCHEDULE` (Python) |
| 环境变量 | `REDIS_URL` | 直接传入 |
| 间隔 | `FB_RATE_LIMIT_MS` (默认 4000ms) | 构造参数 `interval_ms` |

**关键约束**：间隔参数必须一致！

## 降级策略

Go 后端：
- 配置了 `REDIS_URL` → 自动切换为 Redis 分布式模式
- 未配置或 Redis 不可用 → 自动降级为内存限速（仅限 Go 进程内）

Python 端：
- Redis 不可用时需自行处理，建议加上 try/except：

```python
try:
    limiter.acquire()
except redis.ConnectionError:
    print("[WARN] Redis 不可用，跳过限速")
```
