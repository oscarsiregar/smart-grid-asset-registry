package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	healthUp   = "UP"
	healthDown = "DOWN"
)

// HealthCheckResponse is the body of the health endpoint.
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// NewHealthHandler pings the database (and Redis when given) and answers 200
// when every dependency is up, 503 otherwise. Error details are not exposed.
func NewHealthHandler(db *gorm.DB, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		checks := map[string]string{"database": checkDatabase(c, db)}
		if rdb != nil {
			checks["redis"] = healthUp
			if err := rdb.Ping(c.UserContext()).Err(); err != nil {
				checks["redis"] = healthDown
			}
		}

		status, httpStatus := healthUp, fiber.StatusOK
		for _, check := range checks {
			if check != healthUp {
				status, httpStatus = healthDown, fiber.StatusServiceUnavailable
			}
		}

		return c.Status(httpStatus).JSON(HealthCheckResponse{
			Status:    status,
			Timestamp: time.Now().UTC(),
			Checks:    checks,
		})
	}
}

func checkDatabase(c *fiber.Ctx, db *gorm.DB) string {
	sqlDB, err := db.DB()
	if err != nil {
		return healthDown
	}
	if err := sqlDB.PingContext(c.UserContext()); err != nil {
		return healthDown
	}
	return healthUp
}

// RegisterHealthCheck mounts GET /health on router.
func RegisterHealthCheck(router fiber.Router, db *gorm.DB, rdb *redis.Client) {
	router.Get("/health", NewHealthHandler(db, rdb))
}
