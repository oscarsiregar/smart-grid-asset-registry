package utils

import (
	"bytes"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"smart-grid-asset-registry/pkg/gocommon/pkg/apperror"
)

func newErrorApp(t *testing.T, logOutput io.Writer, handlerErr error) *fiber.App {
	t.Helper()

	log := logrus.New()
	log.SetOutput(logOutput)

	app := fiber.New(fiber.Config{ErrorHandler: NewErrorHandler(log)})
	app.Get("/", func(*fiber.Ctx) error { return handlerErr })
	app.Use(NotFoundHandler)
	return app
}

func doGet(t *testing.T, app *fiber.App, path string) (int, string) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func TestNewErrorHandler(t *testing.T) {
	withCode := apperror.New(fiber.StatusUnprocessableEntity, "The file has no data rows")
	withCode.ErrorCode = "IMPORT_FILE_EMPTY"

	serverAppErr := apperror.New(fiber.StatusInternalServerError, "boom")

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
		wantLogged bool
	}{
		{
			name:       "app error with error code",
			err:        withCode,
			wantStatus: fiber.StatusUnprocessableEntity,
			wantBody:   `{"code":422,"status":"error","message":"The file has no data rows","data":{"errorCode":"IMPORT_FILE_EMPTY"}}`,
		},
		{
			name:       "app error without error code",
			err:        apperror.NotFound("Not found"),
			wantStatus: fiber.StatusNotFound,
			wantBody:   `{"code":404,"status":"error","message":"Not found","data":null}`,
		},
		{
			name:       "app error >= 500 is logged",
			err:        serverAppErr,
			wantStatus: fiber.StatusInternalServerError,
			wantBody:   `{"code":500,"status":"error","message":"boom","data":null}`,
			wantLogged: true,
		},
		{
			name:       "fiber error",
			err:        fiber.NewError(fiber.StatusRequestEntityTooLarge, "Request Entity Too Large"),
			wantStatus: fiber.StatusRequestEntityTooLarge,
			wantBody:   `{"code":413,"status":"error","message":"Request Entity Too Large","data":null}`,
		},
		{
			name:       "unknown error is hidden and logged",
			err:        errors.New("pq: connection refused"),
			wantStatus: fiber.StatusInternalServerError,
			wantBody:   `{"code":500,"status":"error","message":"Internal Server Error","data":null}`,
			wantLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			status, body := doGet(t, newErrorApp(t, &logs, tt.err), "/")

			assert.Equal(t, tt.wantStatus, status)
			assert.JSONEq(t, tt.wantBody, body)
			assert.Equal(t, tt.wantLogged, logs.Len() > 0)
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	status, body := doGet(t, newErrorApp(t, io.Discard, nil), "/missing")

	assert.Equal(t, fiber.StatusNotFound, status)
	assert.JSONEq(t, `{"code":404,"status":"error","message":"Endpoint Not Found","data":null}`, body)
}
