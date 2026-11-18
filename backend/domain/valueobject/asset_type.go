package valueobject

import (
	"database/sql/driver"
	"errors"
)

// AssetType is the kind of grid asset, persisted as text. Every enum in this
// service follows the same shape: Parse*, Scan, Value, String, IsValid.
type AssetType string

const (
	AssetTypeSubstation       AssetType = "SUBSTATION"
	AssetTypeTransformer      AssetType = "TRANSFORMER"
	AssetTypeLVBoard          AssetType = "LV_BOARD"
	AssetTypeSwitchboard      AssetType = "SWITCHBOARD"
	AssetTypeSwitchboardPanel AssetType = "SWITCHBOARD_PANEL"
)

var errInvalidAssetType = errors.New("invalid asset type")

func ParseAssetType(value string) (AssetType, error) {
	assetType := AssetType(value)
	if err := assetType.IsValid(); err != nil {
		return "", err
	}
	return assetType, nil
}

func (s *AssetType) Scan(value any) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return errors.New("invalid type for AssetType")
	}

	parsed, err := ParseAssetType(strValue)
	if err != nil {
		return err
	}

	*s = parsed
	return nil
}

func (s AssetType) Value() (driver.Value, error) {
	if err := s.IsValid(); err != nil {
		return nil, err
	}
	return string(s), nil
}

func (s AssetType) String() string {
	return string(s)
}

func (s AssetType) IsValid() error {
	switch s {
	case AssetTypeSubstation,
		AssetTypeTransformer,
		AssetTypeLVBoard,
		AssetTypeSwitchboard,
		AssetTypeSwitchboardPanel:
		return nil
	default:
		return errInvalidAssetType
	}
}

// PermittedParent returns the only asset type allowed as this type's parent
// (spec FR-014). SUBSTATION is a root and has none (ok = false).
func (s AssetType) PermittedParent() (parent AssetType, ok bool) {
	switch s {
	case AssetTypeTransformer, AssetTypeLVBoard, AssetTypeSwitchboard:
		return AssetTypeSubstation, true
	case AssetTypeSwitchboardPanel:
		return AssetTypeSwitchboard, true
	default:
		return "", false
	}
}
