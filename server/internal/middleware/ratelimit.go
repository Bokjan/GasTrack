package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"gastrack/internal/pkg/respond"
)

// RateLimiter 基于 IP 的速率限制器
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit // 每秒允许的请求数
	burst    int        // 突发请求上限
}

// NewRateLimiter 创建速率限制器
// r: 每秒允许的请求数, burst: 突发上限
func NewRateLimiter(r float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(r),
		burst:    burst,
	}
}

// getLimiter 获取或创建 IP 对应的限制器
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
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
