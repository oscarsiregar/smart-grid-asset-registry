package asset_usecase

import (
	"context"
	"errors"
	"strings"
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

func searchFor(text string) repository.FindAssetNodesFilter {
	return repository.FindAssetNodesFilter{Search: pointer(text)}
}

// fiveMatches is what the database returns for "aurora", unsorted: one exact
// match (the substation's name) and four that only contain the text.
func fiveMatches() []*entity.AssetNode {
	return []*entity.AssetNode{
		{AssetID: "PNL-001-1-2", AssetName: "Aurora Panel 2", AssetType: valueobject.AssetTypeSwitchboardPanel},
		{AssetID: "TX-001-1", AssetName: "Aurora Transformer T1", AssetType: valueobject.AssetTypeTransformer},
		{AssetID: "SWB-001-1", AssetName: "Aurora Switchboard A", AssetType: valueobject.AssetTypeSwitchboard},
		{AssetID: "PNL-001-1-10", AssetName: "Aurora Panel 10", AssetType: valueobject.AssetTypeSwitchboardPanel},
		{AssetID: "SS-001", AssetName: "Aurora", AssetType: valueobject.AssetTypeSubstation},
	}
}

// In search order: the exact match first, then type rank, then natural ID.
var fiveMatchesInOrder = []string{"SS-001", "TX-001-1", "SWB-001-1", "PNL-001-1-2", "PNL-001-1-10"}

func TestGetListAsset_Search(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")

	tests := []struct {
		name    string
		input   usecase.GetListAssetInput
		setup   func(d deps)
		check   func(t *testing.T, output usecase.GetListAssetOutput)
		wantErr error
	}{
		{
			name:    "a text that is too long: 400 INVALID_REQUEST, nothing is read",
			input:   usecase.GetListAssetInput{Search: strings.Repeat("a", usecase.MaxSearchLength+1)},
			setup:   func(deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "q must be at most 100 characters"),
		},
		{
			name:    "a page without a search: 400 INVALID_REQUEST, nothing is read",
			input:   usecase.GetListAssetInput{Page: 2},
			setup:   func(deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "page and limit apply to a search only"),
		},
		{
			name:  "first page: matches ordered, the page cut, ancestors of the page only",
			input: usecase.GetListAssetInput{Search: "aurora", Page: 1, Limit: 2},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("aurora")).Return(fiveMatches(), nil)
				d.assetRepo.EXPECT().FindAncestorIDs(mock.Anything, nil, []string{"SS-001", "TX-001-1"}).
					Return(map[string][]string{"SS-001": {}, "TX-001-1": {"SS-001"}}, nil).Once()
			},
			check: func(t *testing.T, output usecase.GetListAssetOutput) {
				assert.Equal(t, fiveMatchesInOrder[:2], nodeIDs(output.Assets))
				assert.Equal(t, 5, output.TotalItems)
				assert.Equal(t, map[string][]string{"SS-001": {}, "TX-001-1": {"SS-001"}}, output.AncestorIDs)
			},
		},
		{
			name:  "last page: the remaining match",
			input: usecase.GetListAssetInput{Search: "aurora", Page: 3, Limit: 2},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("aurora")).Return(fiveMatches(), nil)
				d.assetRepo.EXPECT().FindAncestorIDs(mock.Anything, nil, []string{"PNL-001-1-10"}).
					Return(map[string][]string{"PNL-001-1-10": {"SS-001", "SWB-001-1"}}, nil)
			},
			check: func(t *testing.T, output usecase.GetListAssetOutput) {
				assert.Equal(t, fiveMatchesInOrder[4:], nodeIDs(output.Assets))
				assert.Equal(t, 5, output.TotalItems)
			},
		},
		{
			name:  "a page past the end: no items, the real total, no ancestor query",
			input: usecase.GetListAssetInput{Search: "aurora", Page: 4, Limit: 2},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("aurora")).Return(fiveMatches(), nil)
			},
			check: func(t *testing.T, output usecase.GetListAssetOutput) {
				assert.NotNil(t, output.Assets)
				assert.Empty(t, output.Assets)
				assert.Equal(t, 5, output.TotalItems)
				assert.NotNil(t, output.AncestorIDs)
			},
		},
		{
			name:  "defaults: page 1 of SearchDefaultLimit",
			input: usecase.GetListAssetInput{Search: "aurora"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("aurora")).Return(fiveMatches(), nil)
				d.assetRepo.EXPECT().FindAncestorIDs(mock.Anything, nil, fiveMatchesInOrder).
					Return(map[string][]string{}, nil)
			},
			check: func(t *testing.T, output usecase.GetListAssetOutput) {
				assert.Equal(t, fiveMatchesInOrder, nodeIDs(output.Assets))
			},
		},
		{
			name:  "no matches: empty, total 0, no ancestor query",
			input: usecase.GetListAssetInput{Search: "nothing"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("nothing")).Return([]*entity.AssetNode{}, nil)
			},
			check: func(t *testing.T, output usecase.GetListAssetOutput) {
				assert.NotNil(t, output.Assets)
				assert.Empty(t, output.Assets)
				assert.Zero(t, output.TotalItems)
			},
		},
		{
			name:  "FindNodes error is returned as-is",
			input: usecase.GetListAssetInput{Search: "aurora"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("aurora")).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			name:  "FindAncestorIDs error is returned as-is",
			input: usecase.GetListAssetInput{Search: "aurora"},
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, searchFor("aurora")).Return(fiveMatches(), nil)
				d.assetRepo.EXPECT().FindAncestorIDs(mock.Anything, nil, mock.Anything).Return(nil, repoErr)
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
			tt.check(t, output)
		})
	}
}

func TestGetListAsset_HierarchyHasNoSearchFields(t *testing.T) {
	d := newDeps(t)
	d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, repository.FindAssetNodesFilter{}).Return([]*entity.AssetNode{
		node("SS-001", valueobject.AssetTypeSubstation, 1),
		node("SS-002", valueobject.AssetTypeSubstation, 0),
	}, nil)

	output, err := d.usecase().GetListAsset(context.Background(), usecase.GetListAssetInput{})

	require.NoError(t, err)
	assert.Equal(t, 2, output.TotalItems)
	assert.Nil(t, output.AncestorIDs)
}

func TestGetListAsset_SearchLogging(t *testing.T) {
	ctx := context.Background()

	t.Run("success logs the search length and the total, never the text", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, mock.Anything).Return(fiveMatches(), nil)
		d.assetRepo.EXPECT().FindAncestorIDs(mock.Anything, nil, mock.Anything).Return(map[string][]string{}, nil)

		_, err := u.GetListAsset(ctx, usecase.GetListAssetInput{Search: "aurora"})

		require.NoError(t, err)
		assert.Contains(t, logs.String(), `"msg":"success"`)
		assert.Contains(t, logs.String(), `"searchLength":6`)
		assert.Contains(t, logs.String(), `"total":5`)
		assert.NotContains(t, logs.String(), "aurora")
	})

	t.Run("a technical failure is logged as an error", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()
		d.assetRepo.EXPECT().FindNodes(mock.Anything, nil, mock.Anything).Return(nil, errors.New("db down"))

		_, err := u.GetListAsset(ctx, usecase.GetListAssetInput{Search: "aurora"})

		require.Error(t, err)
		assert.Contains(t, logs.String(), `"level":"error"`)
		assert.Contains(t, logs.String(), "failed to search assets")
	})

	t.Run("an invalid input is not logged as an error", func(t *testing.T) {
		d := newDeps(t)
		u, logs := d.loggingUsecase()

		_, err := u.GetListAsset(ctx, usecase.GetListAssetInput{Page: 2})

		require.Error(t, err)
		assert.NotContains(t, logs.String(), `"level":"error"`)
	})
}
