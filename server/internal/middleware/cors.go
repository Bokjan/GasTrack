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
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Accept-Language"},
		MaxAge:         "86400", // 24 hours
	}
}

// CORS 返回处理跨域请求的中间件
func CORS(cfg CORSConfig) Middleware {
	origins := make(map[string]bool)
	for _, o := range cfg.AllowedOrigins {
		origins[o] = true
	}
	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查 origin 是否允许
			if origins[origin] || origins["*"] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// 处理 preflight 请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
