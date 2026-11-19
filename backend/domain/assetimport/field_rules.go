package assetimport

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"
)

// twoDigitYearBase maps the d/m/yy form's year to 20yy (spec Assumptions).
const twoDigitYearBase = 2000

var (
	// A plain decimal: optional sign, digits with an optional fraction, or a
	// fraction alone. No exponents, thousands separators or units.
	decimalPattern  = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)$`)
	dayFirstPattern = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{2})$`)
	isoDatePattern  = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
)

// checkFields applies the rules that need only the row itself (spec FR-005,
// FR-007 – FR-010) and, when they all pass, builds the row's asset. A row
// whose field count differs from the header's gets MALFORMED_ROW only: its
// values are misaligned, so further field rules would only add noise.
func checkFields(c *candidate, headerFieldCount int) {
	if c.row.FieldCount != headerFieldCount {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.MalformedRow, Args: []any{c.row.FieldCount, headerFieldCount}})
		return
	}

	if c.assetID == "" {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.AssetIDRequired})
	}
	if c.assetType == "" {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.AssetTypeRequired})
	}
	if c.assetName == "" {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.AssetNameRequired})
	}

	assetType, typeErr := valueobject.ParseAssetType(c.assetType)
	if c.assetType != "" && typeErr != nil {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.InvalidAssetType, Args: []any{c.assetType}})
	}

	var status *valueobject.OperationalStatus
	if c.operationalStatus != "" {
		parsed, err := valueobject.ParseOperationalStatus(c.operationalStatus)
		if err != nil {
			c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.InvalidOperationalStatus, Args: []any{c.operationalStatus}})
		} else {
			status = &parsed
		}
	}

	voltage := c.parseNumber(ColumnVoltageKV, c.voltageKV)
	rating := c.parseNumber(ColumnRatingKVA, c.ratingKVA)
	if rating != nil && *rating < 0 {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.NegativeRating, Args: []any{c.ratingKVA}})
	}

	var commissioned *time.Time
	if c.commissionedDate != "" {
		date, ok := parseDate(c.commissionedDate)
		if !ok {
			c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.InvalidDate, Args: []any{c.commissionedDate}})
		} else {
			commissioned = &date
		}
	}

	if len(c.fieldIssues) > 0 {
		return
	}

	c.asset = &entity.Asset{
		AssetID:           c.assetID,
		ParentAssetID:     optional(c.parentAssetID),
		AssetType:         assetType,
		AssetName:         c.assetName,
		OperationalStatus: status,
		VoltageKV:         voltage,
		RatingKVA:         rating,
		Manufacturer:      optional(c.manufacturer),
		Model:             optional(c.model),
		SerialNumber:      optional(c.serialNumber),
		CommissionedDate:  commissioned,
	}
}

// parseNumber returns nil for an empty value and records INVALID_NUMBER for a
// value that is not a plain decimal.
func (c *candidate) parseNumber(column Column, value string) *float64 {
	if value == "" {
		return nil
	}
	if !decimalPattern.MatchString(value) {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.InvalidNumber, Args: []any{string(column), value}})
		return nil
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.InvalidNumber, Args: []any{string(column), value}})
		return nil
	}
	return &number
}

// parseDate accepts d/m/yy (two-digit year = 20yy) and yyyy-mm-dd, and only
// real calendar dates. The result is midnight UTC.
func parseDate(value string) (time.Time, bool) {
	var year, month, day int
	if parts := dayFirstPattern.FindStringSubmatch(value); parts != nil {
		day, month, year = atoi(parts[1]), atoi(parts[2]), twoDigitYearBase+atoi(parts[3])
	} else if parts := isoDatePattern.FindStringSubmatch(value); parts != nil {
		year, month, day = atoi(parts[1]), atoi(parts[2]), atoi(parts[3])
	} else {
		return time.Time{}, false
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || int(date.Month()) != month || date.Day() != day {
		return time.Time{}, false // time.Date normalised an impossible date
	}
	return date, true
}

// atoi is only called on regexp-matched digit strings.
func atoi(digits string) int {
	number, _ := strconv.Atoi(digits)
	return number
}

// checkDuplicateAssetIDs rejects every row whose asset ID appears on more
// than one row (spec FR-006), so the result does not depend on which copy
// comes first. Blank IDs are not compared.
func checkDuplicateAssetIDs(candidates []*candidate) {
	rowsByID := map[string][]int{}
	for _, c := range candidates {
		if c.assetID != "" {
			rowsByID[c.assetID] = append(rowsByID[c.assetID], c.row.Number)
		}
	}

	for _, c := range candidates {
		rows := rowsByID[c.assetID]
		if c.assetID == "" || len(rows) < 2 {
			continue
		}
		c.fieldIssues = append(c.fieldIssues, Issue{Code: errorcode.DuplicateAssetID, Args: []any{c.assetID, joinRows(rows)}})
	}
}

func joinRows(rows []int) string {
	sorted := append([]int(nil), rows...)
	sort.Ints(sorted)

	parts := make([]string, 0, len(sorted))
	for _, row := range sorted {
		parts = append(parts, strconv.Itoa(row))
	}
	return strings.Join(parts, ", ")
}
