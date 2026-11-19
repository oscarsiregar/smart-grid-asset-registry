package fiber

import (
	"bytes"
	"net/http/httptest"
	"testing"

	gofiber "github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
)

func newCommonApp(logs *bytes.Buffer, seenRequestID *string) *gofiber.App {
	log := logrus.New()
	log.SetOutput(logs)
	log.SetFormatter(&logrus.JSONFormatter{})

	app := gofiber.New()
	InstallCommonMiddleware(app, log)
	app.Get("/ok", func(c *gofiber.Ctx) error {
		if id, ok := c.UserContext().Value(logger.RequestIDKey).(string); ok {
			*seenRequestID = id
		}
		return c.SendStatus(gofiber.StatusNoContent)
	})
	app.Get("/fail", func(*gofiber.Ctx) error {
		return gofiber.NewError(gofiber.StatusTeapot, "nope")
	})
	return app
}

func TestInstallCommonMiddleware_GeneratesRequestID(t *testing.T) {
	var logs bytes.Buffer
	var seen string

	resp, err := newCommonApp(&logs, &seen).Test(httptest.NewRequest(gofiber.MethodGet, "/ok", nil))
	require.NoError(t, err)

	id := resp.Header.Get(gofiber.HeaderXRequestID)
	assert.NotEmpty(t, id)
	assert.Equal(t, id, seen, "request id must reach the request context")
	assert.NotEmpty(t, resp.Header.Get(gofiber.HeaderXContentTypeOptions), "security headers must be set")
	assert.Contains(t, logs.String(), `"request_id":"`+id+`"`)
	assert.Contains(t, logs.String(), `"status":204`)
}

func TestInstallCommonMiddleware_ReusesIncomingRequestID(t *testing.T) {
	var logs bytes.Buffer
	var seen string

	req := httptest.NewRequest(gofiber.MethodGet, "/ok", nil)
	req.Header.Set(gofiber.HeaderXRequestID, "req-123")

	resp, err := newCommonApp(&logs, &seen).Test(req)
	require.NoError(t, err)

	assert.Equal(t, "req-123", resp.Header.Get(gofiber.HeaderXRequestID))
	assert.Equal(t, "req-123", seen)
}

func TestInstallCommonMiddleware_LogsErrorStatus(t *testing.T) {
	var logs bytes.Buffer
	var seen string

	resp, err := newCommonApp(&logs, &seen).Test(httptest.NewRequest(gofiber.MethodGet, "/fail", nil))
	require.NoError(t, err)

	assert.Equal(t, gofiber.StatusTeapot, resp.StatusCode)
	assert.Contains(t, logs.String(), `"status":418`)
	assert.NotContains(t, logs.String(), "nope", "access log must not carry error details")
}
