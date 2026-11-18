package utils

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newPingDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(gormlogger.Silent),
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	return db, mock
}

func TestRegisterHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantHealth string
		wantDB     string
	}{
		{name: "database up", wantStatus: fiber.StatusOK, wantHealth: "UP", wantDB: "UP"},
		{name: "database down", pingErr: errors.New("connection refused"), wantStatus: fiber.StatusServiceUnavailable, wantHealth: "DOWN", wantDB: "DOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newPingDB(t)
			mock.ExpectPing().WillReturnError(tt.pingErr)

			app := fiber.New()
			RegisterHealthCheck(app.Group("/api"), db, nil)

			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/health", nil))
			require.NoError(t, err)
			defer resp.Body.Close()

			var body HealthCheckResponse
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantHealth, body.Status)
			assert.Equal(t, map[string]string{"database": tt.wantDB}, body.Checks)
			assert.False(t, body.Timestamp.IsZero())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
