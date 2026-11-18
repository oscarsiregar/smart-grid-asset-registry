package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestLimiterConfig_RequestsUnderThresholdPassThrough(t *testing.T) {
	app := fiber.New()
	app.Use(LimiterConfig())
	app.Get("/x", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLimiterConfig_ExceedingMaxReturnsTooManyRequests(t *testing.T) {
	app := fiber.New()
	app.Use(LimiterConfig())
	// SkipSuccessfulRequests is true, so the handler must return a
	// non-2xx status for the request to count against the limit.
	app.Get("/x", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).SendString("nope")
	})

	for i := 0; i < 20; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	}

	resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Too many requests")
}

func TestRecoverConfig_RecoversFromPanic(t *testing.T) {
	app := fiber.New()
	app.Use(RecoverConfig())
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("boom")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
