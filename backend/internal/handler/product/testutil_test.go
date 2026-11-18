package product

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend-template/domain/entity"
	"backend-template/domain/valueobject"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	pkg_middleware_fiber "repo.smf-indonesia.co.id/package-helper/go-common/pkg/middleware/fiber"
	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/utils"
	"repo.smf-indonesia.co.id/package-helper/go-common/response"
)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

// newTestApp wires the handler like the real router: go-common's error
// handler (so AppError maps to its real HTTP status) plus an
// identity-injecting middleware mirroring Identity()'s c.Locals write.
func newTestApp(h ProductHandler, identity *pkg_middleware_fiber.UserIdentity) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: utils.NewErrorHandler(newTestLogger())})
	if identity != nil {
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(pkg_middleware_fiber.UserIdentityKey, identity)
			return c.Next()
		})
	}
	app.Get("/products", h.GetListProduct)
	app.Post("/products", h.CreateProduct)
	app.Get("/products/:id", h.GetProductDetail)
	app.Put("/products/:id", h.UpdateProduct)
	app.Delete("/products/:id", h.DeleteProduct)
	return app
}

func doRequest(t *testing.T, app *fiber.App, method, target, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func decodeCommon(t *testing.T, body io.Reader) response.Common {
	t.Helper()
	var common response.Common
	require.NoError(t, json.NewDecoder(body).Decode(&common))
	return common
}

func sampleProduct() *entity.Product {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return &entity.Product{
		ID: uuid.New(), Code: "P-001", Name: "Widget", Price: 10,
		Status: valueobject.ProductStatusActive, CreatedAt: now, CreatedByName: "alice", UpdatedAt: now, UpdatedByName: "alice",
	}
}

func identityWith(permissions ...string) *pkg_middleware_fiber.UserIdentity {
	return &pkg_middleware_fiber.UserIdentity{ID: "u-1", Name: "alice", Permissions: permissions}
}
