package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// ClientIPKey context 中存储客户端真实 IP 的 key
const ClientIPKey contextKey = "client_ip"

// ClientIPResolver 根据可信代理列表解析客户端真实 IP。
// 只有当直连对端（RemoteAddr）属于可信代理时，才采信 X-Real-IP / X-Forwarded-For，
// 避免客户端伪造请求头绕过按 IP 的限流。
type ClientIPResolver struct {
	trusted []*net.IPNet
}

// NewClientIPResolver 创建解析器，cidrs 为可信代理网段（也接受单个 IP）
func NewClientIPResolver(cidrs []string) (*ClientIPResolver, error) {
	resolver := &ClientIPResolver{}
	for _, raw := range cidrs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "/") {
			if ip := net.ParseIP(raw); ip != nil && ip.To4() != nil {
				raw += "/32"
			} else {
				raw += "/128"
			}
		}
		_, ipNet, err := net.ParseCIDR(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy %q: %w", raw, err)
		}
		resolver.trusted = append(resolver.trusted, ipNet)
	}
	return resolver, nil
}

func (c *ClientIPResolver) isTrusted(ip net.IP) bool {
	for _, n := range c.trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIP 返回请求的客户端 IP（不含端口）
func (c *ClientIPResolver) ClientIP(r *http.Request) string {
	remote := remoteHost(r.RemoteAddr)
	remoteIP := net.ParseIP(remote)
	if remoteIP == nil || !c.isTrusted(remoteIP) {
		return remote
	}

	// Nginx 以 $remote_addr 覆盖设置 X-Real-IP，来自可信代理时可直接采信
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(realIP) != nil {
		return realIP
	}

	// 否则从右向左跳过可信代理，取第一个不可信地址（左侧部分可被客户端伪造）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		candidate := ""
		for i := len(parts) - 1; i >= 0; i-- {
			p := strings.TrimSpace(parts[i])
			ip := net.ParseIP(p)
			if ip == nil {
				break
			}
			candidate = p
			if !c.isTrusted(ip) {
				break
			}
		}
		if candidate != "" {
			return candidate
		}
	}

	return remote
}

// RealIP 返回将客户端真实 IP 写入 context 的中间件，供日志、限流等后续中间件使用
func RealIP(resolver *ClientIPResolver) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ClientIPKey, resolver.ClientIP(r))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClientIP 获取客户端 IP；未经过 RealIP 中间件时退化为 RemoteAddr 的主机部分
func GetClientIP(r *http.Request) string {
	if ip, ok := r.Context().Value(ClientIPKey).(string); ok && ip != "" {
		return ip
	}
	return remoteHost(r.RemoteAddr)
}

// remoteHost 去掉 RemoteAddr 中的端口
func remoteHost(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
