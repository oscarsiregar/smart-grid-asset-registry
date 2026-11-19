package asset_repository

import (
	"database/sql/driver"
	"strings"
)

const (
	doubleQuote = 0x22 // "
	backslash   = 0x5C // \
)

// textArray sends a []string as ONE PostgreSQL text[] parameter (for
// "= ANY(?)"). GORM would expand a plain slice into one bind parameter per
// element, and a 5 MB file can mention more IDs than PostgreSQL's 65 535
// parameters allow (research R6). Being a driver.Valuer keeps GORM from
// expanding it. Every element is double-quoted, with quotes and backslashes
// escaped, so commas, spaces and braces inside values are safe.
type textArray []string

func (a textArray) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, value := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte(doubleQuote)
		for _, r := range value {
			if r == doubleQuote || r == backslash {
				b.WriteByte(backslash)
			}
			b.WriteRune(r)
		}
		b.WriteByte(doubleQuote)
	}
	b.WriteByte('}')
	return b.String(), nil
}
