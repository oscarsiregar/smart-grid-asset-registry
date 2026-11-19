package valueobject

import (
	"database/sql/driver"
	"errors"
)

// ImportMode decides what a confirm does with rejected rows: PARTIAL (the
// default) saves the valid rows, STRICT saves nothing if any row is rejected.
type ImportMode string

const (
	ImportModePartial ImportMode = "PARTIAL"
	ImportModeStrict  ImportMode = "STRICT"
)

var errInvalidImportMode = errors.New("invalid import mode")

func ParseImportMode(value string) (ImportMode, error) {
	importMode := ImportMode(value)
	if err := importMode.IsValid(); err != nil {
		return "", err
	}
	return importMode, nil
}

func (s *ImportMode) Scan(value any) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return errors.New("invalid type for ImportMode")
	}

	parsed, err := ParseImportMode(strValue)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

func (s ImportMode) Value() (driver.Value, error) {
	if err := s.IsValid(); err != nil {
		return nil, err
	}
	return string(s), nil
}

func (s ImportMode) String() string {
	return string(s)
}

func (s ImportMode) IsValid() error {
	switch s {
	case ImportModePartial,
		ImportModeStrict:
		return nil
	default:
		return errInvalidImportMode
	}
}
