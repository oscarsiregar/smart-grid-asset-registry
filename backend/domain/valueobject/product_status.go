package valueobject

import (
	"database/sql/driver"
	"errors"
)

// ProductStatus is an enum persisted as text. Every enum in this service
// follows the same shape: Parse*, Scan, Value, String, IsValid.
type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "ACTIVE"
	ProductStatusInactive ProductStatus = "INACTIVE"
)

var errInvalidProductStatus = errors.New("invalid product status")

func ParseProductStatus(value string) (ProductStatus, error) {
	status := ProductStatus(value)
	if err := status.IsValid(); err != nil {
		return "", err
	}
	return status, nil
}

func (s *ProductStatus) Scan(value any) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return errors.New("invalid type for ProductStatus")
	}

	parsed, err := ParseProductStatus(strValue)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

func (s ProductStatus) Value() (driver.Value, error) {
	if err := s.IsValid(); err != nil {
		return nil, err
	}
	return string(s), nil
}

func (s ProductStatus) String() string {
	return string(s)
}

func (s ProductStatus) IsValid() error {
	switch s {
	case ProductStatusActive,
		ProductStatusInactive:
		return nil
	default:
		return errInvalidProductStatus
	}
}
