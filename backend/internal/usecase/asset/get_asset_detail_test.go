package asset_usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func panelAsset() *entity.Asset {
	return &entity.Asset{
		AssetID:           "PNL-001-1-01",
		ParentAssetID:     pointer("SWB-001-1"),
		AssetType:         valueobject.AssetTypeSwitchboardPanel,
		AssetName:         "Secret Panel Name",
		OperationalStatus: pointer(valueobject.OperationalStatusMaintenance),
		VoltageKV:         pointer(0.4),
		CommissionedDate:  pointer(time.Date(2010, 9, 11, 0, 0, 0, 0, time.UTC)),
	}
}

func panelPath() []*entity.AssetNode {
	return []*entity.AssetNode{
		node("SS-001", valueobject.AssetTypeSubstation, 5),
		node("SWB-001-1", valueobject.AssetTypeSwitchboard, 4),
		node("PNL-001-1-01", valueobject.AssetTypeSwitchboardPanel, 0),
	}
}

func TestGetAssetDetail(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")

	tests := []struct {
		name    string
		input   usecase.GetAssetDetailInput
		setup   func(d deps)
		check   func(t *testing.T, output usecase.GetAssetDetailOutput)
		wantErr error
	}{
		{
			name:    "blank asset ID: 400 INVALID_REQUEST, nothing is read",
			input:   usecase.GetAssetDetailInput{AssetID: "  "},
			setup:   func(deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "assetId is required"),
		},
		{
			name:  "unknown asset (letter case differs): 404 ASSET_NOT_FOUND",
			input: usecase.GetAssetDetailInput{AssetID: "ss-001"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "ss-001").Return(nil, nil)
			},
			wantErr: apperr.NotFound(errorcode.AssetNotFound, "ss-001"),
		},
		{
			name:  "a panel: the asset, its path substation first, no children",
			input: usecase.GetAssetDetailInput{AssetID: "PNL-001-1-01"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "PNL-001-1-01").Return(panelAsset(), nil)
				d.assetRepo.EXPECT().FindPath(mock.Anything, nil, "PNL-001-1-01").Return(panelPath(), nil)
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("PNL-001-1-01")).Return([]*entity.AssetNode{}, nil)
			},
			check: func(t *testing.T, output usecase.GetAssetDetailOutput) {
				assert.Equal(t, panelAsset(), output.Asset)
				assert.Equal(t, []string{"SS-001", "SWB-001-1", "PNL-001-1-01"}, nodeIDs(output.Path))
				assert.NotNil(t, output.Children)
				assert.Empty(t, output.Children)
			},
		},
		{
			name:  "a switchboard: its panels as children, in natural ID order",
			input: usecase.GetAssetDetailInput{AssetID: "SWB-001-1"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "SWB-001-1").
					Return(&entity.Asset{AssetID: "SWB-001-1", AssetType: valueobject.AssetTypeSwitchboard}, nil)
				d.assetRepo.EXPECT().FindPath(mock.Anything, nil, "SWB-001-1").Return(panelPath()[:2], nil)
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("SWB-001-1")).Return([]*entity.AssetNode{
					node("PNL-001-1-10", valueobject.AssetTypeSwitchboardPanel, 0),
					node("PNL-001-1-02", valueobject.AssetTypeSwitchboardPanel, 0),
					node("PNL-001-1-01", valueobject.AssetTypeSwitchboardPanel, 0),
				}, nil)
			},
			check: func(t *testing.T, output usecase.GetAssetDetailOutput) {
				assert.Equal(t, "SWB-001-1", output.Asset.AssetID)
				assert.Equal(t, []string{"SS-001", "SWB-001-1"}, nodeIDs(output.Path))
				assert.Equal(t, []string{"PNL-001-1-01", "PNL-001-1-02", "PNL-001-1-10"}, nodeIDs(output.Children))
			},
		},
		{
			name:  "FindByAssetID error is returned as-is",
			input: usecase.GetAssetDetailInput{AssetID: "SS-1"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "SS-1").Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			name:  "FindPath error is returned as-is",
			input: usecase.GetAssetDetailInput{AssetID: "PNL-001-1-01"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "PNL-001-1-01").Return(panelAsset(), nil)
				d.assetRepo.EXPECT().FindPath(mock.Anything, nil, "PNL-001-1-01").Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			name:  "FindNodes error is returned as-is",
			input: usecase.GetAssetDetailInput{AssetID: "PNL-001-1-01"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "PNL-001-1-01").Return(panelAsset(), nil)
				d.assetRepo.EXPECT().FindPath(mock.Anything, nil, "PNL-001-1-01").Return(panelPath(), nil)
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("PNL-001-1-01")).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().GetAssetDetail(ctx, tt.input)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, output)
		})
	}
}

func TestGetAssetDetail_Logging(t *testing.T) {
	ctx := context.Background()

	t.Run("success logs the asset ID only, never its attributes", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "PNL-001-1-01").Return(panelAsset(), nil)
		d.assetRepo.EXPECT().FindPath(mock.Anything, nil, "PNL-001-1-01").Return(panelPath(), nil)
		d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, mock.Anything).Return([]*entity.AssetNode{}, nil)

		_, err := u.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "PNL-001-1-01"})

		require.NoError(t, err)
		assert.Contains(t, logs.String(), `"msg":"success"`)
		assert.Contains(t, logs.String(), `"assetId":"PNL-001-1-01"`)
		assert.NotContains(t, logs.String(), "Secret Panel Name")
	})

	t.Run("a technical failure is logged as an error", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "SS-1").Return(nil, errors.New("db down"))

		_, err := u.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "SS-1"})

		require.Error(t, err)
		assert.Contains(t, logs.String(), `"level":"error"`)
		assert.Contains(t, logs.String(), "failed to find asset")
	})

	t.Run("not found is not logged as an error", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "NOPE").Return(nil, nil)

		_, err := u.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "NOPE"})

		require.Error(t, err)
		assert.NotContains(t, logs.String(), `"level":"error"`)
	})
}
