// Package gormtx adapts *gorm.DB to the domain's repository.Tx /
// repository.TransactionManager so usecases never import gorm.
package gormtx

import (
	"context"
	"fmt"

	"backend-template/domain/repository"

	"gorm.io/gorm"
)

type Tx struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Tx {
	return &Tx{DB: db}
}

func (t *Tx) Commit() error {
	return t.DB.Commit().Error
}

func (t *Tx) Rollback() error {
	return t.DB.Rollback().Error
}

// Unwrap returns the *gorm.DB inside trx (nil trx -> nil, nil).
func Unwrap(trx repository.Tx) (*gorm.DB, error) {
	if trx == nil {
		return nil, nil
	}
	t, ok := trx.(*Tx)
	if !ok {
		return nil, fmt.Errorf("gormtx: unsupported Tx type %T", trx)
	}
	return t.DB, nil
}

// Conn picks the connection a repository method must use: the transaction
// when trx is non-nil, otherwise the base db bound to ctx.
func Conn(ctx context.Context, db *gorm.DB, trx repository.Tx) (*gorm.DB, error) {
	if trx == nil {
		return db.WithContext(ctx), nil
	}
	unwrapped, err := Unwrap(trx)
	if err != nil {
		return nil, err
	}
	return unwrapped.WithContext(ctx), nil
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) repository.TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) Begin(ctx context.Context) (repository.Tx, error) {
	trx := tm.db.WithContext(ctx).Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	return New(trx), nil
}

func (tm *transactionManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx repository.Tx) error) error {
	trx, err := tm.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = trx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, trx); err != nil {
		_ = trx.Rollback()
		return err
	}

	return trx.Commit()
}
