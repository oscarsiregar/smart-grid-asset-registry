package errorcode

// Error codes are stable, SCREAMING_SNAKE_CASE identifiers. Every code here
// MUST have a matching entry (ID + EN message) in internal/apperr/apperr.go —
// TestCatalog_EveryErrorCodeHasAMessage enforces that.
const (
	InvalidRequest   = "INVALID_REQUEST"
	PermissionDenied = "PERMISSION_DENIED"

	ProductNotFound          = "PRODUCT_NOT_FOUND"
	ProductCodeAlreadyExists = "PRODUCT_CODE_ALREADY_EXISTS"
)

// All lists every code above. Add new codes here too.
var All = []string{
	InvalidRequest,
	PermissionDenied,
	ProductNotFound,
	ProductCodeAlreadyExists,
}
