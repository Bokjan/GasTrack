// Package apperror 定义了应用层统一的错误类型。
// 所有业务层错误都应使用此包的类型，Handler 层负责将其转为 HTTP 响应。
package apperror

import "fmt"

// AppError 统一的业务错误类型
type AppError struct {
	Code       int    // HTTP 状态码
	BizCode    int    // 业务错误码
	MessageKey string // i18n 消息 key（如 "auth.invalid_credentials"）
	Message    string // 默认英文消息（i18n 查不到时兜底）
	Err        error  // 原始错误（用于日志，不暴露给客户端）
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.BizCode, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.BizCode, e.Message)
}

// Unwrap 支持 errors.Is / errors.As
func (e *AppError) Unwrap() error {
	return e.Err
}

// --- 预定义的错误构造函数 ---

// ErrBadRequest 400 请求参数错误
func ErrBadRequest(messageKey, message string) *AppError {
	return &AppError{Code: 400, BizCode: 4000, MessageKey: messageKey, Message: message}
}

// ErrUnauthorized 401 未认证
func ErrUnauthorized(messageKey, message string) *AppError {
	return &AppError{Code: 401, BizCode: 4010, MessageKey: messageKey, Message: message}
}

// ErrForbidden 403 无权限
func ErrForbidden(messageKey, message string) *AppError {
	return &AppError{Code: 403, BizCode: 4030, MessageKey: messageKey, Message: message}
}

// ErrNotFound 404 资源不存在
func ErrNotFound(messageKey, message string) *AppError {
	return &AppError{Code: 404, BizCode: 4040, MessageKey: messageKey, Message: message}
}

// ErrConflict 409 资源冲突
func ErrConflict(messageKey, message string) *AppError {
	return &AppError{Code: 409, BizCode: 4090, MessageKey: messageKey, Message: message}
}

// ErrValidation 422 校验错误
func ErrValidation(messageKey, message string) *AppError {
	return &AppError{Code: 422, BizCode: 4220, MessageKey: messageKey, Message: message}
}

// ErrInternal 500 服务器内部错误
func ErrInternal(message string, err error) *AppError {
	return &AppError{
		Code:       500,
		BizCode:    5000,
		MessageKey: "error.internal",
		Message:    message,
		Err:        err,
	}
}

// Wrap 包装一个原始错误到 AppError
func Wrap(appErr *AppError, err error) *AppError {
	return &AppError{
		Code:       appErr.Code,
		BizCode:    appErr.BizCode,
		MessageKey: appErr.MessageKey,
		Message:    appErr.Message,
		Err:        err,
	}
}
