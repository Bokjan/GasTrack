// Package decode 提供 HTTP 请求的解析与校验辅助函数。
// 统一处理 JSON 解码、路径参数提取和结构体校验。
package decode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// validate 是全局的校验器实例（线程安全）
var validate = validator.New()

// JSON 从请求体解析 JSON 并校验结构体
// 支持自动校验 struct tag（validate:"required"）
func JSON(r *http.Request, dst interface{}) error {
	// 限制请求体大小（防止大 payload 攻击）
	const maxBodySize = 1 << 20 // 1MB
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		var maxBytesErr *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("malformed JSON at position %d", syntaxErr.Offset)
		case errors.As(err, &unmarshalErr):
			return fmt.Errorf("invalid value for field %q", unmarshalErr.Field)
		case errors.As(err, &maxBytesErr):
			return fmt.Errorf("request body too large (max 1MB)")
		case errors.Is(err, io.EOF):
			return fmt.Errorf("request body is empty")
		default:
			return fmt.Errorf("invalid JSON: %w", err)
		}
	}

	// 校验
	if err := validate.Struct(dst); err != nil {
		return formatValidationErrors(err)
	}

	return nil
}

// PathParam 从 Go 1.22 的 URL 路径参数中提取值
// 示例：对于路由 GET /api/v1/vehicles/{id}，PathParam(r, "id") 返回该 id
func PathParam(r *http.Request, name string) string {
	return r.PathValue(name)
}

// PathParamUUID 提取路径参数并解析为 UUID
func PathParamUUID(r *http.Request, name string) (uuid.UUID, error) {
	raw := r.PathValue(name)
	if raw == "" {
		return uuid.Nil, fmt.Errorf("missing path parameter: %s", name)
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID for %s: %s", name, raw)
	}

	return id, nil
}

// QueryInt 从查询参数中解析整数，带默认值
func QueryInt(r *http.Request, name string, defaultVal int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}

	return val
}

// QueryString 从查询参数中获取字符串，带默认值
func QueryString(r *http.Request, name, defaultVal string) string {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return defaultVal
	}
	return raw
}

// ValidationError 表示校验错误的结构
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// formatValidationErrors 将 validator 的错误转为可读的格式
func formatValidationErrors(err error) error {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}

	var msgs []string
	for _, e := range validationErrs {
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", field))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid email", field))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of: %s", field, e.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid (%s)", field, e.Tag()))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
}

// GetValidationErrors 返回结构化的校验错误列表（用于 API 响应）
func GetValidationErrors(err error) []ValidationError {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return nil
	}

	result := make([]ValidationError, 0, len(validationErrs))
	for _, e := range validationErrs {
		field := strings.ToLower(e.Field())
		var msg string
		switch e.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", field)
		case "email":
			msg = fmt.Sprintf("%s must be a valid email", field)
		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
		case "max":
			msg = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
		default:
			msg = fmt.Sprintf("%s is invalid", field)
		}
		result = append(result, ValidationError{Field: field, Message: msg})
	}
	return result
}
