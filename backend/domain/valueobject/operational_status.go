package valueobject

import (
	"database/sql/driver"
	"errors"
)

// OperationalStatus is an asset's operating state, persisted as text. It is
// optional on an asset: entities hold a *OperationalStatus and NULL means
// "no status" (spec clarification Q4).
type OperationalStatus string

const (
	OperationalStatusInService    OperationalStatus = "IN_SERVICE"
	OperationalStatusMaintenance  OperationalStatus = "MAINTENANCE"
	OperationalStatusOutOfService OperationalStatus = "OUT_OF_SERVICE"
)

var errInvalidOperationalStatus = errors.New("invalid operational status")

func ParseOperationalStatus(value string) (OperationalStatus, error) {
	operationalStatus := OperationalStatus(value)
	if err := operationalStatus.IsValid(); err != nil {
		return "", err
	}
	return operationalStatus, nil
}

func (s *OperationalStatus) Scan(value any) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return errors.New("invalid type for OperationalStatus")
	}

	parsed, err := ParseOperationalStatus(strValue)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

func (s OperationalStatus) Value() (driver.Value, error) {
	if err := s.IsValid(); err != nil {
		return nil, err
	}
	return string(s), nil
}

func (s OperationalStatus) String() string {
	return string(s)
}

func (s OperationalStatus) IsValid() error {
	switch s {
	case OperationalStatusInService,
		OperationalStatusMaintenance,
		OperationalStatusOutOfService:
		return nil
	default:
		return errInvalidOperationalStatus
	}
}
