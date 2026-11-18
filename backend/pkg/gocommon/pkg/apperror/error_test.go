package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		wantCode int
	}{
		{name: "new", err: New(http.StatusTeapot, "msg"), wantCode: http.StatusTeapot},
		{name: "not found", err: NotFound("msg"), wantCode: http.StatusNotFound},
		{name: "bad request", err: BadRequest("msg"), wantCode: http.StatusBadRequest},
		{name: "forbidden", err: Forbidden("msg"), wantCode: http.StatusForbidden},
		{name: "conflict", err: Conflict("msg"), wantCode: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, "msg", tt.err.Message)
			assert.Equal(t, "msg", tt.err.Error())
			assert.Empty(t, tt.err.ErrorCode)
			assert.Nil(t, tt.err.Errors)
		})
	}
}

func TestAppError_ErrorsAs(t *testing.T) {
	appErr := BadRequest("bad")
	appErr.ErrorCode = "INVALID_REQUEST"

	var wrapped error = appErr
	var target *AppError

	assert.True(t, errors.As(wrapped, &target))
	assert.Equal(t, "INVALID_REQUEST", target.ErrorCode)
}
