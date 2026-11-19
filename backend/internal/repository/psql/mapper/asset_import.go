package mapper

import (
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

func ToModelAssetImport(e *entity.AssetImport) *model.AssetImport {
	if e == nil {
		return nil
	}
	return &model.AssetImport{
		ID:            e.ID,
		FileName:      e.FileName,
		Mode:          e.Mode,
		Outcome:       e.Outcome,
		TotalRows:     e.TotalRows,
		ImportedRows:  e.ImportedRows,
		UnchangedRows: e.UnchangedRows,
		RejectedRows:  e.RejectedRows,
		WarningRows:   e.WarningRows,
		CreatedAt:     e.CreatedAt,
		CreatedByName: e.CreatedByName,
		UpdatedAt:     e.UpdatedAt,
		UpdatedByName: e.UpdatedByName,
	}
}

func ToEntityAssetImport(m *model.AssetImport) *entity.AssetImport {
	if m == nil {
		return nil
	}
	return &entity.AssetImport{
		ID:            m.ID,
		FileName:      m.FileName,
		Mode:          m.Mode,
		Outcome:       m.Outcome,
		TotalRows:     m.TotalRows,
		ImportedRows:  m.ImportedRows,
		UnchangedRows: m.UnchangedRows,
		RejectedRows:  m.RejectedRows,
		WarningRows:   m.WarningRows,
		CreatedAt:     m.CreatedAt,
		CreatedByName: m.CreatedByName,
		UpdatedAt:     m.UpdatedAt,
		UpdatedByName: m.UpdatedByName,
	}
}
