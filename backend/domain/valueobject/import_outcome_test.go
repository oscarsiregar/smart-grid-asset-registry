package valueobject

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImportOutcome(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ImportOutcome
		wantErr bool
	}{
		{"validated_only", "VALIDATED_ONLY", ImportOutcomeValidatedOnly, false},
		{"committed", "COMMITTED", ImportOutcomeCommitted, false},
		{"nothing_to_commit", "NOTHING_TO_COMMIT", ImportOutcomeNothingToCommit, false},
		{"strict_blocked", "STRICT_BLOCKED", ImportOutcomeStrictBlocked, false},
		{"lowercase is rejected", "validated_only", "", true},
		{"surrounding whitespace is rejected", " VALIDATED_ONLY ", "", true},
		{"empty", "", "", true},
		{"unknown", "UNKNOWN", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImportOutcome(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestImportOutcome_IsValid(t *testing.T) {
	assert.NoError(t, ImportOutcomeValidatedOnly.IsValid())
	assert.NoError(t, ImportOutcomeCommitted.IsValid())
	assert.NoError(t, ImportOutcomeNothingToCommit.IsValid())
	assert.NoError(t, ImportOutcomeStrictBlocked.IsValid())
	assert.Error(t, ImportOutcome("UNKNOWN").IsValid())
}

func TestImportOutcome_String(t *testing.T) {
	assert.Equal(t, "VALIDATED_ONLY", ImportOutcomeValidatedOnly.String())
}

func TestImportOutcome_Scan(t *testing.T) {
	var s ImportOutcome
	assert.NoError(t, s.Scan("VALIDATED_ONLY"))
	assert.Equal(t, ImportOutcomeValidatedOnly, s)

	var s2 ImportOutcome
	assert.NoError(t, s2.Scan([]byte("STRICT_BLOCKED")))
	assert.Equal(t, ImportOutcomeStrictBlocked, s2)

	var s3 ImportOutcome
	assert.Error(t, s3.Scan(123))

	var s4 ImportOutcome
	assert.Error(t, s4.Scan("UNKNOWN"))
}

func TestImportOutcome_Value(t *testing.T) {
	v, err := ImportOutcomeValidatedOnly.Value()
	assert.NoError(t, err)
	assert.Equal(t, driver.Value("VALIDATED_ONLY"), v)

	_, err = ImportOutcome("UNKNOWN").Value()
	assert.Error(t, err)
}
