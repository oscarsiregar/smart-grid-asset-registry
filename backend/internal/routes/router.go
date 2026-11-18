package routes

import (
	"smart-grid-asset-registry/internal/config"
	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RouteDependencies struct {
	App *fiber.App
	DB  *gorm.DB
	Cfg *config.Config
}

// RegisterRoutes is the only place routes are registered.
//
// Every route is public and lives under /api (constitution, Template
// Exceptions 1 and 5): the product has no users, so the template's
// InternalSecret → Identity → JWTVerifyStrict chain does not apply. If users
// are ever introduced, restore that chain (backend/docs/rules/07-keamanan.md).
func RegisterRoutes(deps *RouteDependencies) {
	api := deps.App.Group("/api")

	// Liveness/readiness probe (pings the DB).
	utils.RegisterHealthCheck(api, deps.DB, nil)
}
