package asset

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// newTestApp wires the handler like the real router, with go-common's error
// handler so AppErrors map to their real status and data.errorCode.
func newTestApp(h AssetHandler) *fiber.App {
	log := logrus.New()
	log.SetOutput(io.Discard)
	app := fiber.New(fiber.Config{ErrorHandler: utils.NewErrorHandler(log)})
	app.Get("/api/assets", h.GetListAsset)
	app.Get("/api/assets/:assetId", h.GetAssetDetail)
	return app
}

func doGet(t *testing.T, app *fiber.App, target string) *http.Response {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, target, nil))
	require.NoError(t, err)
	return resp
}

// doGetRaw sends the target exactly as written, so a malformed escape such
// as %zz reaches the router (httptest.NewRequest would reject it).
func doGetRaw(t *testing.T, app *fiber.App, rawTarget string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.URL = &url.URL{Opaque: rawTarget}
	// app.Test dumps RequestURI, which NewRequest already set to "/".
	req.RequestURI = rawTarget
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

// envelope is the go-common response with a raw data payload.
type envelope struct {
	Code    int             `json:"code"`
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeEnvelope(t *testing.T, body io.Reader) envelope {
	t.Helper()
	var decoded envelope
	require.NoError(t, json.NewDecoder(body).Decode(&decoded))
	return decoded
}

func pointer[T any](value T) *T { return &value }
