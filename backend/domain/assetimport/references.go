package assetimport

import (
	"sort"
	"strings"
)

// ReferencedAssetIDs returns every asset ID and parent asset ID the file
// mentions: trimmed, without blanks, unique and sorted. These are the stored
// assets the rules need to see (research R6).
func (f ParsedFile) ReferencedAssetIDs() []string {
	return uniqueSortedValues(f.Rows, ColumnAssetID, ColumnParentAssetID)
}

// SerialNumbers returns every serial number in the file: trimmed, without
// blanks, unique and sorted.
func (f ParsedFile) SerialNumbers() []string {
	return uniqueSortedValues(f.Rows, ColumnSerialNumber)
}

func uniqueSortedValues(rows []Row, columns ...Column) []string {
	seen := map[string]bool{}
	values := []string{}
	for _, row := range rows {
		for _, column := range columns {
			value := strings.TrimSpace(row.Values[column])
			if value != "" && !seen[value] {
				seen[value] = true
				values = append(values, value)
			}
		}
	}
	sort.Strings(values)
	return values
}
