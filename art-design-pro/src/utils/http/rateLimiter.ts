/**
 * 请求限速队列
 *
 * 确保 FB App 接口请求按最小间隔排队执行，防止被 Facebook 检测为人机。
 * 纯前端内存队列，不依赖 Redis 等服务端组件。
 *
 * @module utils/http/rateLimiter
 */

/** 默认最小请求间隔（毫秒） */
const DEFAULT_INTERVAL_MS = 4000

/** 队列中的请求项 */
interface QueueItem<T> {
  fn: () => Promise<T>
  resolve: (value: T) => void
  reject: (reason?: any) => void
  signal?: { aborted: boolean; addEventListener?: (event: string, fn: () => void) => void }
  url: string
}

/** 在终端输出限速排队信息（仅开发环境） */
function logQueue(url: string, waitMs: number) {
  if (import.meta.env.DEV) {
    console.log(`[RateLimiter] "${url}" 排队等待 ${waitMs}ms`)
  }
}

/** 毫秒级延迟 */
function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * 请求限速器
 *
 * 同一时刻只有一个请求在执行，请求之间至少间隔 minInterval 毫秒。
 * 如果上一个请求还没过间隔时间，后续请求自动排队等待。
 */
class RateLimiter {
  private lastRequestTime = 0
  private readonly minInterval: number
  private queue: QueueItem<any>[] = []
  private processing = false

  constructor(minIntervalMs: number = DEFAULT_INTERVAL_MS) {
    this.minInterval = minIntervalMs
  }

  /**
   * 将请求加入限速队列，返回 Promise
   *
   * @param fn  实际发起请求的函数
   * @param url 请求 URL（用于日志）
   * @param signal 可选的 AbortSignal，用于取消排队中的请求
   */
  schedule<T>(
    fn: () => Promise<T>,
    url: string,
    signal?: { aborted: boolean; addEventListener?: (event: string, fn: () => void) => void }
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      if (signal?.aborted) {
        reject(new DOMException('请求已被取消', 'AbortError'))
        return
      }

      const item: QueueItem<T> = { fn, resolve, reject, signal, url }
      this.queue.push(item)

      // 绑定取消事件：用户取消时从队列中移除
      signal?.addEventListener?.('abort', () => {
        const idx = this.queue.indexOf(item)
        if (idx !== -1) {
          this.queue.splice(idx, 1)
          reject(new DOMException('请求已被取消', 'AbortError'))
        }
      })

      this.process()
    })
  }

  /** 处理队列：按顺序执行，确保请求间隔 >= minInterval */
  private async process() {
    if (this.processing) return
    this.processing = true

    while (this.queue.length > 0) {
      const now = Date.now()
      const elapsed = now - this.lastRequestTime

      // 如果距上次请求不满最小间隔，先等待
      if (this.lastRequestTime > 0 && elapsed < this.minInterval) {
        const waitMs = this.minInterval - elapsed
        logQueue(this.queue[0].url, waitMs)
        await delay(waitMs)
      }

      const item = this.queue.shift()!
      // 如果在排队期间被取消了，跳过
      if (item.signal?.aborted) continue

      this.lastRequestTime = Date.now()

      try {
        const result = await item.fn()
        item.resolve(result)
      } catch (e) {
        item.reject(e)
      }
    }

    this.processing = false
  }

  /** 获取当前队列长度（调试用） */
  get queueLength(): number {
    return this.queue.length
  }

  /** 清空队列并拒绝所有等待中的请求 */
  clear(reason = '队列已清空') {
    while (this.queue.length > 0) {
      const item = this.queue.shift()!
      item.reject(new Error(reason))
    }
  }
}

/**
 * 全局 FB 请求限速器实例
 *
 * 默认间隔 4 秒，可通过环境变量 VITE_FB_RATE_LIMIT_MS 自定义。
 */
const fbInterval = Number(import.meta.env.VITE_FB_RATE_LIMIT_MS) || DEFAULT_INTERVAL_MS
export const fbRateLimiter = new RateLimiter(fbInterval)
