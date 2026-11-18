package pgtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen_SkipsWithoutDatabaseURL(t *testing.T) {
	t.Setenv(EnvDatabaseURL, "")

	var inner *testing.T
	t.Run("open", func(t *testing.T) {
		inner = t
		Open(t)
		t.Error("Open must skip before returning when the URL is unset")
	})

	assert.True(t, inner.Skipped())
}

func TestOpen_MigratesAndCleans(t *testing.T) {
	db := Open(t)

	var tables []string
	err := db.Raw(`SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name IN ('asset', 'asset_import')
		ORDER BY table_name`).Scan(&tables).Error

	assert.NoError(t, err)
	assert.Equal(t, []string{"asset", "asset_import"}, tables)

	var rows int64
	assert.NoError(t, db.Raw("SELECT count(*) FROM asset").Scan(&rows).Error)
	assert.Zero(t, rows)
}
