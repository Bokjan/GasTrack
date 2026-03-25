package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"gastrack/internal/pkg/respond"
)

// contextKey 用于在 context 中存储认证信息
type contextKey string

const (
	// UserIDKey context 中存储用户 ID 的 key
	UserIDKey contextKey = "user_id"
	// UserLocaleKey context 中存储用户语言偏好的 key
	UserLocaleKey contextKey = "user_locale"
)

// Auth 返回 JWT 认证中间件
// 从 Authorization: Bearer <token> 中提取并验证 JWT
func Auth(jwtSecret string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从 Header 提取 Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respond.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respond.Unauthorized(w, "invalid authorization format")
				return
			}

			tokenStr := parts[1]

			// 解析并验证 JWT
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				respond.Unauthorized(w, "invalid or expired token")
				return
			}

			// 提取 claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respond.Unauthorized(w, "invalid token claims")
				return
			}

			// 提取用户 ID
			sub, ok := claims["sub"].(string)
			if !ok {
				respond.Unauthorized(w, "invalid token subject")
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				respond.Unauthorized(w, "invalid user ID in token")
				return
			}

			// 将用户 ID 存入 context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// 也可以存入用户语言偏好（从 Accept-Language 或 token 中获取）
			locale := r.Header.Get("Accept-Language")
			if locale == "" {
				locale = "en-US"
			}
			ctx = context.WithValue(ctx, UserLocaleKey, locale)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID 从 context 中获取当前登录用户的 ID
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// GetUserLocale 从 context 中获取用户语言偏好
func GetUserLocale(ctx context.Context) string {
	locale, ok := ctx.Value(UserLocaleKey).(string)
	if !ok {
		return "en-US"
	}
	return locale
}
