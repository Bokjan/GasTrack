package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig CORS 中间件配置
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         string // preflight 缓存时间
}

// DefaultCORSConfig 返回默认 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Accept-Language"},
		MaxAge:         "86400", // 24 hours
	}
}

// CORS 返回处理跨域请求的中间件
// API 使用 Bearer Token 认证、不依赖 Cookie，因此不下发 Allow-Credentials；
// 配置 "*" 时返回通配符而非反射请求的 Origin。
func CORS(cfg CORSConfig) Middleware {
	origins := make(map[string]bool)
	allowAll := false
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			allowAll = true
		}
		origins[o] = true
	}
	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			// 响应内容随 Origin 变化，告知缓存按 Origin 区分
			w.Header().Add("Vary", "Origin")

			allowed := origin != "" && (allowAll || origins[origin])
			if allowed {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
				w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)
			}

			// 处理 preflight 请求
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
