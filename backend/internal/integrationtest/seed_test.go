// Package integrationtest runs repositories and usecases against a real
// PostgreSQL database (constitution, Template Exception 11). Every test is
// skipped unless TEST_DATABASE_URL is set (see internal/testsupport/pgtest).
package integrationtest

import (
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/repository/psql/mapper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var seedTime = time.Date(2026, time.November, 19, 9, 0, 0, 0, time.UTC)

// seedImport stores an import record that seeded assets can reference.
func seedImport(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()

	record := &entity.AssetImport{
		ID:            uuid.New(),
		FileName:      "seed.csv",
		Mode:          valueobject.ImportModePartial,
		Outcome:       valueobject.ImportOutcomeCommitted,
		CreatedAt:     seedTime,
		CreatedByName: "system",
		UpdatedAt:     seedTime,
		UpdatedByName: "system",
	}
	require.NoError(t, db.Create(mapper.ToModelAssetImport(record)).Error)
	return record.ID
}

// seedAssets stores assets directly, parents first, bypassing the usecases.
func seedAssets(t *testing.T, db *gorm.DB, importID uuid.UUID, assets ...*entity.Asset) {
	t.Helper()

	for _, asset := range assets {
		asset.ID = uuid.New()
		asset.ImportID = importID
		asset.CreatedAt, asset.UpdatedAt = seedTime, seedTime
		asset.CreatedByName, asset.UpdatedByName = "system", "system"
		require.NoError(t, db.Create(mapper.ToModelAsset(asset)).Error)
	}
}

func pointer[T any](value T) *T {
	return &value
}
