package integrationtest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/bootstrap"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/routes"
	"smart-grid-asset-registry/internal/testsupport/pgtest"
	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newRouterApp wires the real dependencies and routes, as main does.
func newRouterApp(db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: utils.NewErrorHandler(quietLogger())})
	routes.RegisterRoutes(bootstrap.SetupDependencies(app, db, nil, quietLogger()))
	return app
}

type routerResponse struct {
	status int
	body   struct {
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
}

// getRaw sends the target exactly as written (already percent-encoded).
func getRaw(t *testing.T, app *fiber.App, rawTarget string) routerResponse {
	t.Helper()
	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.URL = &url.URL{Opaque: rawTarget}
	// app.Test dumps RequestURI, which NewRequest already set to "/".
	req.RequestURI = rawTarget
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var result routerResponse
	result.status = resp.StatusCode
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.body))
	return result
}

// TestAssetExplorer_RouterRoundTripsEncodedIDs: an ID with a slash and a
// space round-trips through the address, and IDs match exactly (US3,
// FR-017, FR-020).
func TestAssetExplorer_RouterRoundTripsEncodedIDs(t *testing.T) {
	db := pgtest.Open(t)
	importUsecase := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	importSample(t, importUsecase)
	commit(t, importUsecase, valueobject.ImportModeStrict,
		"SS/9 A,,SUBSTATION,Slash Substation,IN_SERVICE,22,,,,,",
		"TX/9 A-1,SS/9 A,TRANSFORMER,Slash Transformer,MAINTENANCE,22,1000,,,,",
	)
	app := newRouterApp(db)

	t.Run("the detail of SS/9 A is reached through SS%2F9%20A", func(t *testing.T) {
		response := getRaw(t, app, "/api/assets/SS%2F9%20A")

		require.Equal(t, http.StatusOK, response.status)
		var detail struct {
			Asset struct {
				AssetID string `json:"assetId"`
			} `json:"asset"`
			Children []struct {
				AssetID string `json:"assetId"`
			} `json:"children"`
		}
		require.NoError(t, json.Unmarshal(response.body.Data, &detail))
		assert.Equal(t, "SS/9 A", detail.Asset.AssetID)
		require.Len(t, detail.Children, 1)
		assert.Equal(t, "TX/9 A-1", detail.Children[0].AssetID)
	})

	t.Run("its children list is reached through parentAssetId=SS%2F9%20A", func(t *testing.T) {
		response := getRaw(t, app, "/api/assets?parentAssetId=SS%2F9%20A")

		require.Equal(t, http.StatusOK, response.status)
		var list struct {
			Items []struct {
				AssetID string `json:"assetId"`
			} `json:"items"`
		}
		require.NoError(t, json.Unmarshal(response.body.Data, &list))
		require.Len(t, list.Items, 1)
		assert.Equal(t, "TX/9 A-1", list.Items[0].AssetID)
	})

	notFound := []struct {
		name    string
		target  string
		message string
	}{
		{"letter case differs from SS-001", "/api/assets/ss-001", "No asset with ID ss-001 exists."},
		{"an unknown ID", "/api/assets/NOPE", "No asset with ID NOPE exists."},
		{"an unknown parent", "/api/assets?parentAssetId=NOPE", "No asset with ID NOPE exists."},
	}
	for _, tc := range notFound {
		t.Run(tc.name+" is 404 ASSET_NOT_FOUND", func(t *testing.T) {
			response := getRaw(t, app, tc.target)

			require.Equal(t, http.StatusNotFound, response.status)
			assert.Equal(t, tc.message, response.body.Message)
			assert.JSONEq(t, `{"errorCode":"ASSET_NOT_FOUND"}`, string(response.body.Data))
		})
	}

	t.Run("a malformed escape is 400 INVALID_REQUEST", func(t *testing.T) {
		response := getRaw(t, app, "/api/assets/%zz")

		require.Equal(t, http.StatusBadRequest, response.status)
		assert.JSONEq(t, `{"errorCode":"INVALID_REQUEST"}`, string(response.body.Data))
	})
}
