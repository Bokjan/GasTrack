package apperror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := ErrBadRequest("test.key", "bad request")
		assert.Equal(t, "[4000] bad request", err.Error())
	})

	t.Run("with wrapped error", func(t *testing.T) {
		inner := errors.New("db connection failed")
		err := ErrInternal("internal error", inner)
		assert.Equal(t, "[5000] internal error: db connection failed", err.Error())
	})
}

func TestAppError_Unwrap(t *testing.T) {
	t.Run("nil inner error", func(t *testing.T) {
		err := ErrNotFound("test.key", "not found")
		assert.Nil(t, err.Unwrap())
	})

	t.Run("with inner error", func(t *testing.T) {
		inner := errors.New("original")
		err := ErrInternal("wrapped", inner)
		assert.Equal(t, inner, err.Unwrap())
		assert.True(t, errors.Is(err, inner))
	})
}

func TestErrBadRequest(t *testing.T) {
	err := ErrBadRequest("validation.required", "field is required")
	assert.Equal(t, 400, err.Code)
	assert.Equal(t, 4000, err.BizCode)
	assert.Equal(t, "validation.required", err.MessageKey)
	assert.Equal(t, "field is required", err.Message)
	assert.Nil(t, err.Err)
}

func TestErrUnauthorized(t *testing.T) {
	err := ErrUnauthorized("auth.invalid", "invalid token")
	assert.Equal(t, 401, err.Code)
	assert.Equal(t, 4010, err.BizCode)
	assert.Equal(t, "auth.invalid", err.MessageKey)
	assert.Equal(t, "invalid token", err.Message)
}

func TestErrForbidden(t *testing.T) {
	err := ErrForbidden("auth.forbidden", "no permission")
	assert.Equal(t, 403, err.Code)
	assert.Equal(t, 4030, err.BizCode)
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("resource.not_found", "not found")
	assert.Equal(t, 404, err.Code)
	assert.Equal(t, 4040, err.BizCode)
}

func TestErrConflict(t *testing.T) {
	err := ErrConflict("resource.conflict", "already exists")
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, 4090, err.BizCode)
}

func TestErrValidation(t *testing.T) {
	err := ErrValidation("validation.error", "invalid input")
	assert.Equal(t, 422, err.Code)
	assert.Equal(t, 4220, err.BizCode)
}

func TestErrInternal(t *testing.T) {
	inner := errors.New("db error")
	err := ErrInternal("something went wrong", inner)
	assert.Equal(t, 500, err.Code)
	assert.Equal(t, 5000, err.BizCode)
	assert.Equal(t, "error.internal", err.MessageKey)
	assert.Equal(t, "something went wrong", err.Message)
	assert.Equal(t, inner, err.Err)
}

func TestWrap(t *testing.T) {
	original := ErrNotFound("user.not_found", "user not found")
	inner := errors.New("record not found")
	wrapped := Wrap(original, inner)

	assert.Equal(t, original.Code, wrapped.Code)
	assert.Equal(t, original.BizCode, wrapped.BizCode)
	assert.Equal(t, original.MessageKey, wrapped.MessageKey)
	assert.Equal(t, original.Message, wrapped.Message)
	assert.Equal(t, inner, wrapped.Err)
	// original should not be mutated
	assert.Nil(t, original.Err)
}

func TestAppError_ImplementsErrorInterface(t *testing.T) {
	var err error = ErrBadRequest("test", "test")
	assert.NotNil(t, err)

	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 400, appErr.Code)
}
