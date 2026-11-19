package gormtx

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return gdb, mock
}

type otherTx struct{}

func (otherTx) Commit() error   { return nil }
func (otherTx) Rollback() error { return nil }

func TestUnwrap(t *testing.T) {
	gdb, _ := newMockDB(t)

	db, err := Unwrap(nil)
	assert.NoError(t, err)
	assert.Nil(t, db)

	db, err = Unwrap(New(gdb))
	assert.NoError(t, err)
	assert.Same(t, gdb, db)

	_, err = Unwrap(otherTx{})
	assert.Error(t, err)
}

func TestConn(t *testing.T) {
	gdb, _ := newMockDB(t)
	ctx := context.Background()

	db, err := Conn(ctx, gdb, nil)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	db, err = Conn(ctx, gdb, New(gdb))
	assert.NoError(t, err)
	assert.NotNil(t, db)

	_, err = Conn(ctx, gdb, otherTx{})
	assert.Error(t, err)
}

func TestWithinTx(t *testing.T) {
	ctx := context.Background()

	t.Run("commits when fn succeeds", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectCommit()

		err := NewTransactionManager(gdb).WithinTx(ctx, func(ctx context.Context, tx repository.Tx) error { return nil })
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rolls back and returns fn's error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectRollback()

		fnErr := errors.New("boom")
		err := NewTransactionManager(gdb).WithinTx(ctx, func(ctx context.Context, tx repository.Tx) error { return fnErr })
		assert.ErrorIs(t, err, fnErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rolls back and re-panics", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectRollback()

		assert.Panics(t, func() {
			_ = NewTransactionManager(gdb).WithinTx(ctx, func(ctx context.Context, tx repository.Tx) error { panic("boom") })
		})
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns the begin error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		beginErr := errors.New("begin failed")
		mock.ExpectBegin().WillReturnError(beginErr)

		err := NewTransactionManager(gdb).WithinTx(ctx, func(ctx context.Context, tx repository.Tx) error { return nil })
		assert.ErrorIs(t, err, beginErr)
	})
}
