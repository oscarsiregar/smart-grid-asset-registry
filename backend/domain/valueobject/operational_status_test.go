package valueobject

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOperationalStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    OperationalStatus
		wantErr bool
	}{
		{"in_service", "IN_SERVICE", OperationalStatusInService, false},
		{"maintenance", "MAINTENANCE", OperationalStatusMaintenance, false},
		{"out_of_service", "OUT_OF_SERVICE", OperationalStatusOutOfService, false},
		{"lowercase is rejected", "in_service", "", true},
		{"surrounding whitespace is rejected", " IN_SERVICE ", "", true},
		{"empty", "", "", true},
		{"unknown", "UNKNOWN", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOperationalStatus(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOperationalStatus_IsValid(t *testing.T) {
	assert.NoError(t, OperationalStatusInService.IsValid())
	assert.NoError(t, OperationalStatusMaintenance.IsValid())
	assert.NoError(t, OperationalStatusOutOfService.IsValid())
	assert.Error(t, OperationalStatus("UNKNOWN").IsValid())
}

func TestOperationalStatus_String(t *testing.T) {
	assert.Equal(t, "IN_SERVICE", OperationalStatusInService.String())
}

func TestOperationalStatus_Scan(t *testing.T) {
	var s OperationalStatus
	assert.NoError(t, s.Scan("IN_SERVICE"))
	assert.Equal(t, OperationalStatusInService, s)

	var s2 OperationalStatus
	assert.NoError(t, s2.Scan([]byte("OUT_OF_SERVICE")))
	assert.Equal(t, OperationalStatusOutOfService, s2)

	var s3 OperationalStatus
	assert.Error(t, s3.Scan(123))

	var s4 OperationalStatus
	assert.Error(t, s4.Scan("UNKNOWN"))
}

func TestOperationalStatus_Value(t *testing.T) {
	v, err := OperationalStatusInService.Value()
	assert.NoError(t, err)
	assert.Equal(t, driver.Value("IN_SERVICE"), v)

	_, err = OperationalStatus("UNKNOWN").Value()
	assert.Error(t, err)
}
