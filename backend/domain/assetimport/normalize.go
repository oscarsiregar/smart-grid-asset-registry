package assetimport

import (
	"strings"

	"smart-grid-asset-registry/domain/entity"
)

// candidate is one data row on its way through the rules. Every value is
// trimmed; a whitespace-only value is empty. Inner whitespace is kept.
type candidate struct {
	row Row

	assetID           string
	parentAssetID     string
	assetType         string
	assetName         string
	operationalStatus string
	voltageKV         string
	ratingKVA         string
	manufacturer      string
	model             string
	serialNumber      string
	commissionedDate  string

	fieldIssues []Issue
	// hierarchyIssues holds every rule broken against other rows or the
	// store: the parent rules and ASSET_ID_CONFLICT.
	hierarchyIssues []Issue

	// unchanged is set when the row's asset ID is stored with the same
	// values (FR-018); such a row is neither imported nor rejected.
	unchanged bool

	// asset is built by checkFields when the row's own fields are valid.
	asset *entity.Asset
}

func newCandidate(row Row) *candidate {
	value := func(column Column) string { return strings.TrimSpace(row.Values[column]) }

	return &candidate{
		row:               row,
		assetID:           value(ColumnAssetID),
		parentAssetID:     value(ColumnParentAssetID),
		assetType:         value(ColumnAssetType),
		assetName:         value(ColumnAssetName),
		operationalStatus: value(ColumnOperationalStatus),
		voltageKV:         value(ColumnVoltageKV),
		ratingKVA:         value(ColumnRatingKVA),
		manufacturer:      value(ColumnManufacturer),
		model:             value(ColumnModel),
		serialNumber:      value(ColumnSerialNumber),
		commissionedDate:  value(ColumnCommissionedDate),
	}
}

func (c *candidate) rejected() bool {
	return len(c.fieldIssues) > 0 || len(c.hierarchyIssues) > 0
}

func (c *candidate) fieldsValid() bool {
	return len(c.fieldIssues) == 0
}

// optional returns nil for an empty value.
func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
