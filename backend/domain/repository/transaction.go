package repository

import (
	"context"
)

// Tx is an opaque transaction handle. Usecases pass it to repository methods
// without knowing it wraps a *gorm.DB (see internal/repository/psql/gormtx).
type Tx interface {
	Commit() error
	Rollback() error
}

type TransactionManager interface {
	Begin(ctx context.Context) (Tx, error)
	// WithinTx runs fn inside a transaction: rollback when fn returns an
	// error (or panics), commit otherwise.
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}
