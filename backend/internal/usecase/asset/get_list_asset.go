package asset_usecase

import (
	"context"
	"unicode/utf8"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"

	"github.com/sirupsen/logrus"
)

// GetListAsset returns either the hierarchy of feature 002 — the complete,
// ordered top level (ParentAssetID nil) or the children of one asset
// (Template Exception 13) — or, with a search text, one page of matching
// assets (feature 003). An unknown parent is 404 ASSET_NOT_FOUND; a stored
// asset without children is an empty list.
func (u assetUsecase) GetListAsset(ctx context.Context, input usecase.GetListAssetInput) (usecase.GetListAssetOutput, error) {
	log := logger.WithCtx(ctx, u.log).WithField("usecase", "GetListAsset")

	input.SetDefault()
	if err := input.Validate(); err != nil {
		return usecase.GetListAssetOutput{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	if input.Search != "" {
		return u.searchAssets(ctx, log, input)
	}

	var noTransaction repository.Tx
	nodes, err := u.assetRepository.FindNodes(ctx, noTransaction, repository.FindAssetNodesFilter{ParentAssetID: input.ParentAssetID})
	if err != nil {
		log.WithError(err).Error("failed to find asset nodes")
		return usecase.GetListAssetOutput{}, err
	}

	// An empty child list needs one more read to tell "no children" from
	// "no such parent".
	if len(nodes) == 0 && input.ParentAssetID != nil {
		parent, err := u.assetRepository.FindByAssetID(ctx, noTransaction, *input.ParentAssetID)
		if err != nil {
			log.WithError(err).Error("failed to find parent asset")
			return usecase.GetListAssetOutput{}, err
		}
		if parent == nil {
			return usecase.GetListAssetOutput{}, apperr.NotFound(errorcode.AssetNotFound, *input.ParentAssetID)
		}
	}

	entity.SortAssetNodes(nodes)

	log.WithField("parentAssetId", parentLogValue(input.ParentAssetID)).WithField("count", len(nodes)).Info("success")
	return usecase.GetListAssetOutput{Assets: nodes, TotalItems: len(nodes)}, nil
}

// searchAssets loads every match, orders it (exact matches first, then the
// explorer's natural order) and cuts the requested page here, because that
// order cannot be expressed in SQL (Template Exception 14). Only the page's
// ancestors are read. The log carries the text's length, never the text.
func (u assetUsecase) searchAssets(ctx context.Context, log *logrus.Entry, input usecase.GetListAssetInput) (usecase.GetListAssetOutput, error) {
	var noTransaction repository.Tx
	matches, err := u.assetRepository.FindNodes(ctx, noTransaction, repository.FindAssetNodesFilter{
		Search:    &input.Search,
		AssetType: input.AssetType,
	})
	if err != nil {
		log.WithError(err).Error("failed to search assets")
		return usecase.GetListAssetOutput{}, err
	}

	entity.SortSearchMatches(matches, input.Search)
	page := pageOf(matches, input.Page, input.Limit)

	ancestors := map[string][]string{}
	if len(page) > 0 {
		ancestors, err = u.assetRepository.FindAncestorIDs(ctx, noTransaction, nodeAssetIDs(page))
		if err != nil {
			log.WithError(err).Error("failed to find ancestors of search results")
			return usecase.GetListAssetOutput{}, err
		}
	}

	log.WithFields(logrus.Fields{
		"searchLength": utf8.RuneCountInString(input.Search),
		"total":        len(matches),
		"page":         input.Page,
		"count":        len(page),
	}).Info("success")
	return usecase.GetListAssetOutput{Assets: page, TotalItems: len(matches), AncestorIDs: ancestors}, nil
}
