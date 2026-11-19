package mapper

import (
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

func ToModelAsset(e *entity.Asset) *model.Asset {
	if e == nil {
		return nil
	}
	return &model.Asset{
		ID:                e.ID,
		AssetID:           e.AssetID,
		ParentAssetID:     e.ParentAssetID,
		AssetType:         e.AssetType,
		AssetName:         e.AssetName,
		OperationalStatus: e.OperationalStatus,
		VoltageKV:         e.VoltageKV,
		RatingKVA:         e.RatingKVA,
		Manufacturer:      e.Manufacturer,
		Model:             e.Model,
		SerialNumber:      e.SerialNumber,
		CommissionedDate:  e.CommissionedDate,
		ImportID:          e.ImportID,
		CreatedAt:         e.CreatedAt,
		CreatedByName:     e.CreatedByName,
		UpdatedAt:         e.UpdatedAt,
		UpdatedByName:     e.UpdatedByName,
	}
}

func ToEntityAsset(m *model.Asset) *entity.Asset {
	if m == nil {
		return nil
	}
	return &entity.Asset{
		ID:                m.ID,
		AssetID:           m.AssetID,
		ParentAssetID:     m.ParentAssetID,
		AssetType:         m.AssetType,
		AssetName:         m.AssetName,
		OperationalStatus: m.OperationalStatus,
		VoltageKV:         m.VoltageKV,
		RatingKVA:         m.RatingKVA,
		Manufacturer:      m.Manufacturer,
		Model:             m.Model,
		SerialNumber:      m.SerialNumber,
		CommissionedDate:  m.CommissionedDate,
		ImportID:          m.ImportID,
		CreatedAt:         m.CreatedAt,
		CreatedByName:     m.CreatedByName,
		UpdatedAt:         m.UpdatedAt,
		UpdatedByName:     m.UpdatedByName,
	}
}

// ToEntityAssets always returns a non-nil slice.
func ToEntityAssets(models []*model.Asset) []*entity.Asset {
	assets := make([]*entity.Asset, 0, len(models))
	for _, m := range models {
		assets = append(assets, ToEntityAsset(m))
	}
	return assets
}

// ToModelAssets always returns a non-nil slice.
func ToModelAssets(entities []*entity.Asset) []*model.Asset {
	models := make([]*model.Asset, 0, len(entities))
	for _, e := range entities {
		models = append(models, ToModelAsset(e))
	}
	return models
}
