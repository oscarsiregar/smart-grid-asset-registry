package assetimport

import (
	"sort"
	"strings"

	"smart-grid-asset-registry/domain/errorcode"
)

// serialWarnings warns on every accepted row whose serial number is also used
// by another accepted row or by a stored asset with a different asset ID
// (spec FR-020). Rejected rows neither get nor cause warnings, and a warning
// never changes whether a row is accepted.
func serialWarnings(candidates []*candidate, stored StoredAssets) []Warning {
	idsBySerial := map[string][]string{}
	for _, c := range candidates {
		if !c.rejected() && c.serialNumber != "" {
			idsBySerial[c.serialNumber] = append(idsBySerial[c.serialNumber], c.assetID)
		}
	}

	var warnings []Warning
	for _, c := range candidates {
		if c.rejected() || c.serialNumber == "" {
			continue
		}

		others := otherAssetIDs(c.assetID, idsBySerial[c.serialNumber], stored.BySerial[c.serialNumber])
		if len(others) == 0 {
			continue
		}
		warnings = append(warnings, Warning{
			Row:     c.row.Number,
			AssetID: c.assetID,
			Issue:   Issue{Code: errorcode.DuplicateSerialNumber, Args: []any{c.serialNumber, strings.Join(others, ", ")}},
		})
	}

	sort.Slice(warnings, func(i, j int) bool { return warnings[i].Row < warnings[j].Row })
	return warnings
}

// otherAssetIDs returns the sorted, unique IDs in both lists except ownID.
func otherAssetIDs(ownID string, inFile, inStore []string) []string {
	seen := map[string]bool{ownID: true}
	var others []string
	for _, id := range append(append([]string(nil), inFile...), inStore...) {
		if !seen[id] {
			seen[id] = true
			others = append(others, id)
		}
	}
	sort.Strings(others)
	return others
}
