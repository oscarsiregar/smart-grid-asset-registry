package mapper

import (
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAssetImportMapper_RoundTrip(t *testing.T) {
	now := time.Date(2026, time.November, 19, 10, 0, 0, 0, time.UTC)
	e := &entity.AssetImport{
		ID:            uuid.New(),
		FileName:      "grid_assets.csv",
		Mode:          valueobject.ImportModePartial,
		Outcome:       valueobject.ImportOutcomeCommitted,
		TotalRows:     222,
		ImportedRows:  205,
		UnchangedRows: 0,
		RejectedRows:  17,
		WarningRows:   2,
		CreatedAt:     now,
		CreatedByName: "system",
		UpdatedAt:     now,
		UpdatedByName: "system",
	}

	m := ToModelAssetImport(e)
	assert.Equal(t, &model.AssetImport{
		ID: e.ID, FileName: "grid_assets.csv", Mode: valueobject.ImportModePartial, Outcome: valueobject.ImportOutcomeCommitted,
		TotalRows: 222, ImportedRows: 205, UnchangedRows: 0, RejectedRows: 17, WarningRows: 2,
		CreatedAt: now, CreatedByName: "system", UpdatedAt: now, UpdatedByName: "system",
	}, m)

	assert.Equal(t, e, ToEntityAssetImport(m))
}

func TestAssetImportMapper_Nil(t *testing.T) {
	assert.Nil(t, ToModelAssetImport(nil))
	assert.Nil(t, ToEntityAssetImport(nil))
}
