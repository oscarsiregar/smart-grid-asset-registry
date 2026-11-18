// Package service holds interfaces for OUTBOUND integrations (other
// microservices, object storage, message brokers, ...).
//
// Rules:
//   - Only the interface + its Input/Output structs live here; the HTTP/SDK
//     implementation lives in internal/service/<name>/ (one file per call).
//   - Usecases depend on these interfaces, never on the implementation.
//   - Add the package to .mockery.yaml once it has its first interface.
//
// Example:
//
//	type UserManagementService interface {
//	    GetUserByID(ctx context.Context, input GetUserByIDInput) (*GetUserByIDOutput, error)
//	}
package service
