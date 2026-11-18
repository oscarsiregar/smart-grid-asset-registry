package valueobject

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImportMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ImportMode
		wantErr bool
	}{
		{"partial", "PARTIAL", ImportModePartial, false},
		{"strict", "STRICT", ImportModeStrict, false},
		{"lowercase is rejected", "partial", "", true},
		{"surrounding whitespace is rejected", " PARTIAL ", "", true},
		{"empty", "", "", true},
		{"unknown", "UNKNOWN", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImportMode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestImportMode_IsValid(t *testing.T) {
	assert.NoError(t, ImportModePartial.IsValid())
	assert.NoError(t, ImportModeStrict.IsValid())
	assert.Error(t, ImportMode("UNKNOWN").IsValid())
}

func TestImportMode_String(t *testing.T) {
	assert.Equal(t, "PARTIAL", ImportModePartial.String())
}

func TestImportMode_Scan(t *testing.T) {
	var s ImportMode
	assert.NoError(t, s.Scan("PARTIAL"))
	assert.Equal(t, ImportModePartial, s)

	var s2 ImportMode
	assert.NoError(t, s2.Scan([]byte("STRICT")))
	assert.Equal(t, ImportModeStrict, s2)

	var s3 ImportMode
	assert.Error(t, s3.Scan(123))

	var s4 ImportMode
	assert.Error(t, s4.Scan("UNKNOWN"))
}

func TestImportMode_Value(t *testing.T) {
	v, err := ImportModePartial.Value()
	assert.NoError(t, err)
	assert.Equal(t, driver.Value("PARTIAL"), v)

	_, err = ImportMode("UNKNOWN").Value()
	assert.Error(t, err)
}
