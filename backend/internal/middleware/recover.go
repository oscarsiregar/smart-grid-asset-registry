package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// RecoverConfig turns a panic in any handler into a 500 for that request
// only — the process keeps serving other requests.
func RecoverConfig() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
	})
}
