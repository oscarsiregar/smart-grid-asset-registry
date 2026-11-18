package valueobject

// Permission strings are issued by the identity service and arrive on every
// request via the BFF (X-User-Permissions / JWT). Format: "<module>.<action>".
// One constant per endpoint; keep them grouped by module and sorted by action.
const (
	PermissionCreateProduct    = "product.create"
	PermissionDeleteProduct    = "product.delete"
	PermissionGetListProduct   = "product.get_list"
	PermissionGetProductDetail = "product.get_detail"
	PermissionUpdateProduct    = "product.update"
)
