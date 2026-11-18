package migrations

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_EmbedsEveryMigrationPair(t *testing.T) {
	names, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{
		"000001_create_asset_import.up.sql",
		"000001_create_asset_import.down.sql",
		"000002_create_asset.up.sql",
		"000002_create_asset.down.sql",
	}, names)
}

func TestPostgresURL_EscapesCredentials(t *testing.T) {
	got := PostgresURL("postgres", "p@ss:w/rd", "db", 5432, "asset_registry")
	assert.Equal(t, "postgres://postgres:p%40ss%3Aw%2Frd@db:5432/asset_registry?sslmode=disable", got)
}

func TestToDriverURL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"postgres scheme", "postgres://u:p@h:5432/d?sslmode=disable", "pgx5://u:p@h:5432/d?sslmode=disable", false},
		{"postgresql scheme", "postgresql://u@h/d", "pgx5://u@h/d", false},
		{"other scheme", "mysql://u@h/d", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toDriverURL(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUp_RejectsNonPostgresURL(t *testing.T) {
	assert.Error(t, Up("mysql://u@h/d"))
}
