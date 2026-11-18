package response

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func call(t *testing.T, handler fiber.Handler) (int, string) {
	t.Helper()

	app := fiber.New()
	app.Get("/", handler)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(body)
}

func TestSuccess(t *testing.T) {
	status, body := call(t, func(c *fiber.Ctx) error {
		return Success(c, fiber.StatusCreated, "Successfully create thing", map[string]string{"id": "1"})
	})

	assert.Equal(t, fiber.StatusCreated, status)
	assert.JSONEq(t, `{"code":201,"status":"success","message":"Successfully create thing","data":{"id":"1"}}`, body)
}

func TestError(t *testing.T) {
	status, body := call(t, func(c *fiber.Ctx) error {
		return Error(c, fiber.StatusNotFound, "Not found", nil)
	})

	assert.Equal(t, fiber.StatusNotFound, status)
	assert.JSONEq(t, `{"code":404,"status":"error","message":"Not found","data":null}`, body)
}

func TestDataKeyIsAlwaysPresent(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		want string
	}{
		{name: "nil", data: nil, want: `{"code":200,"status":"success","message":"ok","data":null}`},
		{name: "empty slice", data: []string{}, want: `{"code":200,"status":"success","message":"ok","data":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, body := call(t, func(c *fiber.Ctx) error {
				return Success(c, fiber.StatusOK, "ok", tt.data)
			})
			assert.JSONEq(t, tt.want, body)
		})
	}
}

func TestJSON_StatusFollowsCode(t *testing.T) {
	_, body := call(t, func(c *fiber.Ctx) error {
		return JSON(c, fiber.StatusBadRequest, "bad", "x")
	})

	assert.JSONEq(t, `{"code":400,"status":"error","message":"bad","data":"x"}`, body)
}
