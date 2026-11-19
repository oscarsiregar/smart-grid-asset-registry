package valueobject

import (
	"database/sql/driver"
	"errors"
)

// ImportOutcome is the result of validating or confirming an import file.
// VALIDATED_ONLY is only ever returned, never stored; the other values are
// stored on asset_import records.
type ImportOutcome string

const (
	ImportOutcomeValidatedOnly   ImportOutcome = "VALIDATED_ONLY"
	ImportOutcomeCommitted       ImportOutcome = "COMMITTED"
	ImportOutcomeNothingToCommit ImportOutcome = "NOTHING_TO_COMMIT"
	ImportOutcomeStrictBlocked   ImportOutcome = "STRICT_BLOCKED"
)

var errInvalidImportOutcome = errors.New("invalid import outcome")

func ParseImportOutcome(value string) (ImportOutcome, error) {
	importOutcome := ImportOutcome(value)
	if err := importOutcome.IsValid(); err != nil {
		return "", err
	}
	return importOutcome, nil
}

func (s *ImportOutcome) Scan(value any) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return errors.New("invalid type for ImportOutcome")
	}

	parsed, err := ParseImportOutcome(strValue)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

func (s ImportOutcome) Value() (driver.Value, error) {
	if err := s.IsValid(); err != nil {
		return nil, err
	}
	return string(s), nil
}

func (s ImportOutcome) String() string {
	return string(s)
}

func (s ImportOutcome) IsValid() error {
	switch s {
	case ImportOutcomeValidatedOnly,
		ImportOutcomeCommitted,
		ImportOutcomeNothingToCommit,
		ImportOutcomeStrictBlocked:
		return nil
	default:
		return errInvalidImportOutcome
	}
}
