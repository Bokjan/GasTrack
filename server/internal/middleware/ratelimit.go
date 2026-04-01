package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"gastrack/internal/pkg/respond"
)

// ipLimiter 带最后访问时间的限流器
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter 基于 IP 的速率限制器（带自动清理过期条目）
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rate     rate.Limit // 每秒允许的请求数
	burst    int        // 突发请求上限
}

// NewRateLimiter 创建速率限制器
// r: 每秒允许的请求数, burst: 突发上限
func NewRateLimiter(r float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     rate.Limit(r),
		burst:    burst,
	}

	// 启动后台 goroutine 定期清理过期 IP 条目，防止内存泄漏
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop 每 5 分钟清理 10 分钟未活跃的 IP 限流器
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, il := range rl.limiters {
			if il.lastSeen.Before(cutoff) {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getLimiter 获取或创建 IP 对应的限制器
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	il, exists := rl.limiters[ip]
	if !exists {
		il = &ipLimiter{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = il
	} else {
		il.lastSeen = time.Now()
	}

	return il.limiter
}

// RateLimit 返回 IP 级别限流中间件
func RateLimit(requestsPerSecond float64, burst int) Middleware {
	rl := NewRateLimiter(requestsPerSecond, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			// 检查 X-Forwarded-For（如果经过反向代理）
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				respond.Error(w, http.StatusTooManyRequests, 4290, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
