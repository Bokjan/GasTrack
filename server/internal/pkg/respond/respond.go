// Package respond 提供统一的 HTTP JSON 响应辅助函数。
// 所有 API 响应格式保持一致，便于前端解析。
package respond

import (
	"encoding/json"
	"net/http"
)

// Response 统一的 API 响应结构
type Response struct {
	Code    int         `json:"code"`              // 业务状态码，0 表示成功
	Message string      `json:"message"`           // 消息描述
	Data    interface{} `json:"data,omitempty"`    // 响应数据
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    int         `json:"code"`              // 业务错误码
	Message string      `json:"message"`           // 错误消息
	Errors  interface{} `json:"errors,omitempty"` // 详细错误（校验错误等）
}

// PagedResponse 分页响应结构
type PagedResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    PageMeta    `json:"meta"`
}

// PageMeta 分页元信息
type PageMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// JSON 写入成功的 JSON 响应
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// OK 200 成功响应
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created 201 创建成功响应
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent 204 无内容响应（删除成功等场景）
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Paged 带分页的响应
func Paged(w http.ResponseWriter, data interface{}, page, pageSize int, total int64) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	resp := PagedResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		Meta: PageMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// Error 写入错误 JSON 响应
func Error(w http.ResponseWriter, statusCode int, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Code:    code,
		Message: message,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// ValidationError 写入校验错误响应（422）
func ValidationError(w http.ResponseWriter, errors interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)

	resp := ErrorResponse{
		Code:    4220,
		Message: "validation_error",
		Errors:  errors,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// BadRequest 400 错误请求
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, 4000, message)
}

// Unauthorized 401 未认证
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, 4010, message)
}

// Forbidden 403 无权限
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, 4030, message)
}

// NotFound 404 资源不存在
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, 4040, message)
}

// InternalError 500 服务器内部错误
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, 5000, message)
}
