package assetimport

import (
	"sort"

	"smart-grid-asset-registry/domain/entity"
)

// Validate applies every import rule to a parsed file against the stored
// assets it mentions and returns a deterministic report: the same rows in
// any order give an identical report (spec FR-017).
//
// Order of the rules: field rules per row and duplicate IDs over the file,
// then the hierarchy rules on rows whose fields are valid, then serial-number
// warnings on the accepted rows.
func Validate(file ParsedFile, stored StoredAssets) Report {
	candidates := make([]*candidate, 0, len(file.Rows))
	for _, row := range file.Rows {
		candidates = append(candidates, newCandidate(row))
	}

	for _, c := range candidates {
		checkFields(c, file.HeaderFieldCount)
	}
	checkDuplicateAssetIDs(candidates)
	checkHierarchy(candidates, stored)
	warnings := serialWarnings(candidates, stored)

	return buildReport(candidates, warnings)
}

func buildReport(candidates []*candidate, warnings []Warning) Report {
	report := Report{
		Accepted:      []*entity.Asset{},
		UnchangedRows: []int{},
		Rejected:      []RejectedRow{},
		Warnings:      warnings,
	}
	if report.Warnings == nil {
		report.Warnings = []Warning{}
	}

	for _, c := range candidates {
		if !c.rejected() {
			report.Accepted = append(report.Accepted, c.asset)
			continue
		}
		report.Rejected = append(report.Rejected, RejectedRow{
			Row:     c.row.Number,
			AssetID: optional(c.assetID),
			Errors:  append(append([]Issue(nil), c.fieldIssues...), c.hierarchyIssues...),
		})
	}

	sort.Slice(report.Rejected, func(i, j int) bool { return report.Rejected[i].Row < report.Rejected[j].Row })
	sortParentsFirst(report.Accepted)

	warningRows := map[int]bool{}
	for _, warning := range report.Warnings {
		warningRows[warning.Row] = true
	}

	report.Counts = Counts{
		Total:       len(candidates),
		ToImport:    len(report.Accepted),
		Unchanged:   len(report.UnchangedRows),
		Rejected:    len(report.Rejected),
		WarningRows: len(warningRows),
	}
	return report
}

// sortParentsFirst orders assets by depth within the list (an asset whose
// parent is not in the list has depth 0), then by asset ID, so a batched
// insert never references a parent that is not stored yet.
func sortParentsFirst(assets []*entity.Asset) {
	byID := make(map[string]*entity.Asset, len(assets))
	for _, asset := range assets {
		byID[asset.AssetID] = asset
	}

	depth := make(map[string]int, len(assets))
	var depthOf func(asset *entity.Asset) int
	depthOf = func(asset *entity.Asset) int {
		if known, ok := depth[asset.AssetID]; ok {
			return known
		}
		result := 0
		if asset.ParentAssetID != nil {
			if parent, ok := byID[*asset.ParentAssetID]; ok {
				result = depthOf(parent) + 1
			}
		}
		depth[asset.AssetID] = result
		return result
	}

	sort.Slice(assets, func(i, j int) bool {
		di, dj := depthOf(assets[i]), depthOf(assets[j])
		if di != dj {
			return di < dj
		}
		return assets[i].AssetID < assets[j].AssetID
	})
}
