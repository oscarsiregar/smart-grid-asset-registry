// Package querysafety sanitizes user-controlled values that cannot be bound
// as SQL parameters (ORDER BY column/direction).
package querysafety

import (
	"regexp"
	"strings"
)

var columnIdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// SafeOrderClause returns "<column> <ASC|DESC>". A sort value that is not a
// plain identifier, or not in allowedSorts (when given), falls back to
// defaultSort; an order other than ASC/DESC falls back to defaultOrder.
func SafeOrderClause(sort, order, defaultSort, defaultOrder string, allowedSorts ...string) string {
	if !columnIdentifierPattern.MatchString(sort) || (len(allowedSorts) > 0 && !contains(allowedSorts, sort)) {
		sort = defaultSort
	}

	normalizedOrder := strings.ToUpper(order)
	if normalizedOrder != "ASC" && normalizedOrder != "DESC" {
		normalizedOrder = defaultOrder
	}

	return sort + " " + normalizedOrder
}

func contains(values []string, v string) bool {
	for _, value := range values {
		if value == v {
			return true
		}
	}
	return false
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// ContainsPattern turns user input into a safe "%...%" LIKE/ILIKE pattern:
// %, _ and \ typed by the user are matched literally, not as wildcards.
func ContainsPattern(q string) string {
	return "%" + likeEscaper.Replace(q) + "%"
}
