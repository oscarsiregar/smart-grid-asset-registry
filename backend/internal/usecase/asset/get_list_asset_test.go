package asset_usecase

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func childrenOf(id string) repository.FindAssetNodesFilter {
	return repository.FindAssetNodesFilter{ParentAssetID: pointer(id)}
}

func TestGetListAsset(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	topLevel := repository.FindAssetNodesFilter{}

	tests := []struct {
		name    string
		input   usecase.GetListAssetInput
		setup   func(d deps)
		wantIDs []string
		wantErr error
	}{
		{
			name:  "top level: every substation, in natural ID order",
			input: usecase.GetListAssetInput{},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, topLevel).Return([]*entity.AssetNode{
					node("SS-T01", valueobject.AssetTypeSubstation, 1),
					node("SS-010", valueobject.AssetTypeSubstation, 2),
					node("SS-002", valueobject.AssetTypeSubstation, 0),
				}, nil)
			},
			wantIDs: []string{"SS-002", "SS-010", "SS-T01"},
		},
		{
			name:  "children of a stored asset: grouped by type, then by ID",
			input: usecase.GetListAssetInput{ParentAssetID: pointer("SS-001")},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("SS-001")).Return([]*entity.AssetNode{
					node("SWB-001-1", valueobject.AssetTypeSwitchboard, 4),
					node("TX-001-2", valueobject.AssetTypeTransformer, 0),
					node("LVB-001-1", valueobject.AssetTypeLVBoard, 0),
					node("TX-001-1", valueobject.AssetTypeTransformer, 0),
				}, nil)
			},
			wantIDs: []string{"TX-001-1", "TX-001-2", "LVB-001-1", "SWB-001-1"},
		},
		{
			name:  "no children and the parent exists: empty list",
			input: usecase.GetListAssetInput{ParentAssetID: pointer("PNL-1")},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("PNL-1")).Return([]*entity.AssetNode{}, nil)
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "PNL-1").Return(&entity.Asset{AssetID: "PNL-1"}, nil)
			},
			wantIDs: []string{},
		},
		{
			name:  "no children and the parent does not exist: 404 ASSET_NOT_FOUND",
			input: usecase.GetListAssetInput{ParentAssetID: pointer("NOPE")},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("NOPE")).Return([]*entity.AssetNode{}, nil)
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "NOPE").Return(nil, nil)
			},
			wantErr: apperr.NotFound(errorcode.AssetNotFound, "NOPE"),
		},
		{
			name:  "empty registry: empty top level, no error",
			input: usecase.GetListAssetInput{},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, topLevel).Return([]*entity.AssetNode{}, nil)
			},
			wantIDs: []string{},
		},
		{
			name:  "FindNodes error is returned as-is",
			input: usecase.GetListAssetInput{},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, topLevel).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			name:  "FindByAssetID error is returned as-is",
			input: usecase.GetListAssetInput{ParentAssetID: pointer("SS-1")},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, childrenOf("SS-1")).Return([]*entity.AssetNode{}, nil)
				d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "SS-1").Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().GetListAsset(ctx, tt.input)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, output.Assets)
			assert.Equal(t, tt.wantIDs, nodeIDs(output.Assets))
		})
	}
}

func TestGetListAsset_Logging(t *testing.T) {
	ctx := context.Background()

	t.Run("success logs the parent and the count, no row contents", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, mock.Anything).Return([]*entity.AssetNode{
			node("TX-1", valueobject.AssetTypeTransformer, 0),
		}, nil)

		_, err := u.GetListAsset(ctx, usecase.GetListAssetInput{ParentAssetID: pointer("SS-1")})

		require.NoError(t, err)
		assert.Contains(t, logs.String(), `"msg":"success"`)
		assert.Contains(t, logs.String(), `"parentAssetId":"SS-1"`)
		assert.Contains(t, logs.String(), `"count":1`)
		assert.NotContains(t, logs.String(), "Name TX-1")
	})

	t.Run("a technical failure is logged as an error", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, mock.Anything).Return(nil, errors.New("db down"))

		_, err := u.GetListAsset(ctx, usecase.GetListAssetInput{})

		require.Error(t, err)
		assert.Contains(t, logs.String(), `"level":"error"`)
		assert.Contains(t, logs.String(), "failed to find asset nodes")
	})

	t.Run("an unknown parent is not logged as an error", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, mock.Anything).Return([]*entity.AssetNode{}, nil)
		d.assetRepo.EXPECT().FindByAssetID(mock.Anything, nil, "NOPE").Return(nil, nil)

		_, err := u.GetListAsset(ctx, usecase.GetListAssetInput{ParentAssetID: pointer("NOPE")})

		require.Error(t, err)
		assert.NotContains(t, logs.String(), `"level":"error"`)
	})
}
