package assetimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsedFile_References(t *testing.T) {
	file := mustParse(t,
		withSerial("TX-1", "SS-9", "TRANSFORMER", "S-2"),
		withSerial(" SS-1 ", "", "SUBSTATION", " S-1 "),
		withSerial("TX-2", "SS-1", "TRANSFORMER", ""),
		withSerial("", "SS-9", "TRANSFORMER", "S-2"),
	)

	assert.Equal(t, []string{"SS-1", "SS-9", "TX-1", "TX-2"}, file.ReferencedAssetIDs(),
		"asset and parent IDs, trimmed, unique, sorted, no blanks")
	assert.Equal(t, []string{"S-1", "S-2"}, file.SerialNumbers())
}

func TestParsedFile_ReferencesOfAnEmptyFileAreEmpty(t *testing.T) {
	assert.Empty(t, ParsedFile{}.ReferencedAssetIDs())
	assert.NotNil(t, ParsedFile{}.ReferencedAssetIDs())
	assert.NotNil(t, ParsedFile{}.SerialNumbers())
}
