package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"repo.smf-indonesia.co.id/package-helper/go-common/response"
)

// LimiterConfig throttles clients that keep FAILING (non-2xx): successful
// requests are not counted, so normal traffic is never limited.
func LimiterConfig() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,
		Expiration: 15 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).
				JSON(response.Common{
					Code:    fiber.StatusTooManyRequests,
					Status:  "error",
					Message: "Too many requests, please try again later",
				})
		},
		SkipSuccessfulRequests: true,
	})
}
