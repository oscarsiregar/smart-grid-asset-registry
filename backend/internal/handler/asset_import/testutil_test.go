package asset_import

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

// newTestApp wires the handler like the real router, with go-common's error
// handler so AppErrors map to their real status and data.errorCode.
func newTestApp(h AssetImportHandler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: utils.NewErrorHandler(newTestLogger())})
	app.Post("/api/imports/validate", h.ValidateAssetImport)
	app.Post("/api/imports", h.CreateAssetImport)
	return app
}

// multipartPart is one form part; a part with a file name is a file.
type multipartPart struct {
	name     string
	fileName string
	content  string
}

func doMultipart(t *testing.T, app *fiber.App, target string, parts ...multipartPart) *http.Response {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		if part.fileName != "" {
			fileWriter, err := writer.CreateFormFile(part.name, part.fileName)
			require.NoError(t, err)
			_, err = fileWriter.Write([]byte(part.content))
			require.NoError(t, err)
			continue
		}
		require.NoError(t, writer.WriteField(part.name, part.content))
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(fiber.MethodPost, target, &body)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())
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

func filePart(content string) multipartPart {
	return multipartPart{name: "file", fileName: "grid_assets.csv", content: content}
}

func modePart(mode string) multipartPart {
	return multipartPart{name: "mode", content: mode}
}
