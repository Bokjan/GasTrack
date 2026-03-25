// Package middleware 提供 HTTP 中间件。
// 中间件通过标准库 http.Handler 接口实现链式组合。
package middleware

import "net/http"

// Middleware 定义中间件函数类型
type Middleware func(http.Handler) http.Handler

// Chain 将多个中间件按顺序组合成一个
// 执行顺序与传入顺序一致：Chain(A, B, C)(handler) → A(B(C(handler)))
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// ChainFunc 将多个中间件组合后应用到 HandlerFunc
func ChainFunc(handler http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(handler)
}
