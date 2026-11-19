package assetimport

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const goldenFile = "testdata/grid_assets.golden.json"

type goldenRejected struct {
	Row     int      `json:"row"`
	AssetID *string  `json:"assetId"`
	Codes   []string `json:"codes"`
}

type goldenWarning struct {
	Row     int    `json:"row"`
	AssetID string `json:"assetId"`
	Code    string `json:"code"`
}

type golden struct {
	Counts         Counts           `json:"counts"`
	AcceptedByType map[string]int   `json:"acceptedByType"`
	Rejected       []goldenRejected `json:"rejected"`
	Warnings       []goldenWarning  `json:"warnings"`
}

// TestGolden_SampleFile pins the result of the supplied sample CSV against an
// empty store (constitution, principle IV). The expected values in
// testdata/grid_assets.golden.json were derived by hand from the spec's rules
// (research R12) and checked row by row against the file; they are not
// generated from this implementation.
func TestGolden_SampleFile(t *testing.T) {
	raw, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	var want golden
	require.NoError(t, json.Unmarshal(raw, &want))

	report := Validate(parseSample(t), emptyStore())

	got := golden{
		Counts:         report.Counts,
		AcceptedByType: map[string]int{},
		Rejected:       make([]goldenRejected, 0, len(report.Rejected)),
		Warnings:       make([]goldenWarning, 0, len(report.Warnings)),
	}
	for _, asset := range report.Accepted {
		got.AcceptedByType[asset.AssetType.String()]++
	}
	for _, rejected := range report.Rejected {
		codes := make([]string, 0, len(rejected.Errors))
		for _, issue := range rejected.Errors {
			codes = append(codes, issue.Code)
		}
		got.Rejected = append(got.Rejected, goldenRejected{Row: rejected.Row, AssetID: rejected.AssetID, Codes: codes})
	}
	for _, warning := range report.Warnings {
		got.Warnings = append(got.Warnings, goldenWarning{Row: warning.Row, AssetID: warning.AssetID, Code: warning.Issue.Code})
	}

	assert.Equal(t, want, got)
}
