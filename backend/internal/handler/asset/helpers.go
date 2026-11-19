package asset

import (
	"time"

	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
)

func toAssetNodeResponse(node *entity.AssetNode) dto.AssetNodeResponse {
	return dto.AssetNodeResponse{
		AssetID:           node.AssetID,
		AssetName:         node.AssetName,
		AssetType:         node.AssetType.String(),
		OperationalStatus: statusText(node.OperationalStatus),
		ChildCount:        node.ChildCount,
	}
}

// toAssetNodeResponses always returns a non-nil slice, so JSON shows [].
func toAssetNodeResponses(nodes []*entity.AssetNode) []dto.AssetNodeResponse {
	responses := make([]dto.AssetNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		responses = append(responses, toAssetNodeResponse(node))
	}
	return responses
}

func toAssetListResponse(output usecase.GetListAssetOutput) dto.AssetListResponse {
	items := toAssetNodeResponses(output.Assets)
	return dto.AssetListResponse{Items: items, TotalItems: len(items)}
}

func toAssetResponse(asset *entity.Asset) dto.AssetResponse {
	return dto.AssetResponse{
		AssetID:           asset.AssetID,
		ParentAssetID:     asset.ParentAssetID,
		AssetType:         asset.AssetType.String(),
		AssetName:         asset.AssetName,
		OperationalStatus: statusText(asset.OperationalStatus),
		VoltageKV:         asset.VoltageKV,
		RatingKVA:         asset.RatingKVA,
		Manufacturer:      asset.Manufacturer,
		Model:             asset.Model,
		SerialNumber:      asset.SerialNumber,
		CommissionedDate:  dateText(asset.CommissionedDate),
	}
}

func toAssetDetailResponse(output usecase.GetAssetDetailOutput) dto.AssetDetailResponse {
	return dto.AssetDetailResponse{
		Asset:    toAssetResponse(output.Asset),
		Path:     toAssetNodeResponses(output.Path),
		Children: toAssetNodeResponses(output.Children),
	}
}

// statusText: nil (no status recorded) stays null in JSON.
func statusText(status *valueobject.OperationalStatus) *string {
	if status == nil {
		return nil
	}
	text := status.String()
	return &text
}

// dateText formats a stored date as YYYY-MM-DD; nil stays null.
func dateText(date *time.Time) *string {
	if date == nil {
		return nil
	}
	text := date.Format(time.DateOnly)
	return &text
}
