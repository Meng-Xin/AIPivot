package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindowLimiter 基于 Redis ZSET 的滑动窗口限流器。
// 适用于 Widget 访客维度限流（如：每访客每分钟最多 10 次发送）。
//
// 原理：以 timestamp(ms) 为 ZSET member 存入窗口内的请求，每次请求：
//  1. ZREMRANGEBYSCORE 清掉窗口外的旧记录
//  2. ZCARD 统计窗口内请求数
//  3. 未超限则 ZADD 当前请求并设 TTL
type SlidingWindowLimiter struct {
	redis   *redis.Client
	window  time.Duration // 窗口大小
	limit   int64         // 窗口内最大请求数
}

// NewSlidingWindowLimiter 构造限流器。
//   - window: 窗口长度（如 time.Minute）
//   - limit:  窗口内允许的最大请求数（0 表示禁用限流，永远放行）
func NewSlidingWindowLimiter(rdb *redis.Client, window time.Duration, limit int64) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{redis: rdb, window: window, limit: limit}
}

// Allow 判断 key（如 "widget:msg:<visitorId>"）是否在窗口内未超限。
// limit=0 时放行；Redis 故障时 fail-open（不阻断业务）。
func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) bool {
	if l.limit <= 0 {
		return true
	}
	now := time.Now().UnixMilli()
	windowStart := now - l.window.Milliseconds()

	// 用 Lua 保证「清理 + 计数 + 写入」原子性
	// 1. ZREMRANGEBYSCORE key -inf windowStart
	// 2. ZCARD key
	// 3. 若 count < limit：ZADD key now now；EXPIRE key window+buffer；返回 1
	//    否则返回 0
	script := redis.NewScript(`
		redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
		local cnt = redis.call('ZCARD', KEYS[1])
		if cnt < tonumber(ARGV[2]) then
			redis.call('ZADD', KEYS[1], ARGV[3], ARGV[3])
			redis.call('PEXPIRE', KEYS[1], ARGV[4])
			return 1
		end
		return 0
	`)

	ttlMs := l.window.Milliseconds() + 1000 // 多留 1s 缓冲
	res, err := script.Run(ctx, l.redis, []string{key},
		strconv.FormatInt(windowStart, 10),
		strconv.FormatInt(l.limit, 10),
		strconv.FormatInt(now, 10),
		strconv.FormatInt(ttlMs, 10),
	).Int64()
	if err != nil {
		// Redis 故障 fail-open
		return true
	}
	return res == 1
}
