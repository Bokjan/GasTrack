package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"

	"gastrack/internal/pkg/respond"
)

// Recovery 返回 panic 恢复中间件
// 捕获 handler 中的 panic，记录堆栈信息并返回 500 错误
func Recovery(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 记录 panic 堆栈
					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("stack", string(debug.Stack())),
					)

					respond.InternalError(w, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
