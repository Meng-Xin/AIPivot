// Package ratelimit 提供基于 Redis 的日粒度 Token 配额限流器。
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"aipivot/internal/shared/errorx"

	"github.com/redis/go-redis/v9"
)

// TokenLimiter 基于 Redis 的租户日 Token 配额限流器。
// Redis key 格式: rl:tokens:{tenantID}:{YYYY-MM-DD}，TTL 48 小时。
type TokenLimiter struct {
	redis *redis.Client
	limit int64 // 每日 token 上限，0 = 不限制
}

// NewTokenLimiter 构造限流器；limit=0 时禁用限流（所有请求放行）。
func NewTokenLimiter(rdb *redis.Client, limit int64) *TokenLimiter {
	return &TokenLimiter{redis: rdb, limit: limit}
}

// Check 检查租户当日 token 配额是否已耗尽。
// limit=0 时始终返回 nil；Redis 故障时 fail-open（放行）。
func (l *TokenLimiter) Check(ctx context.Context, tenantID int64) error {
	if l.limit <= 0 {
		return nil
	}
	cur, err := l.redis.Get(ctx, dayKey(tenantID)).Int64()
	if err != nil && err != redis.Nil {
		// Redis 故障时不阻断业务
		return nil
	}
	if cur >= l.limit {
		return errorx.NewBusinessError(errorx.CodeTokenExceeded, "今日 Token 配额已耗尽，请明日再试")
	}
	return nil
}

// Incr 异步增加租户当日已用 token 数（fire-and-forget）。
// TTL 设为 48h 确保跨午夜的请求不会误判。
func (l *TokenLimiter) Incr(ctx context.Context, tenantID int64, amount int) {
	if amount <= 0 || l.limit <= 0 {
		return
	}
	key := dayKey(tenantID)
	pipe := l.redis.Pipeline()
	pipe.IncrBy(ctx, key, int64(amount))
	pipe.Expire(ctx, key, 48*time.Hour)
	_, _ = pipe.Exec(ctx)
}

// DailyUsage 返回租户当日已使用的 token 数（Redis 查询失败时返回 0）。
func (l *TokenLimiter) DailyUsage(ctx context.Context, tenantID int64) int64 {
	cur, err := l.redis.Get(ctx, dayKey(tenantID)).Int64()
	if err != nil {
		return 0
	}
	return cur
}

func dayKey(tenantID int64) string {
	return fmt.Sprintf("rl:tokens:%d:%s", tenantID, time.Now().Format("2006-01-02"))
}
