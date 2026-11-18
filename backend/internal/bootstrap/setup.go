// Package bootstrap is the composition root: the ONLY place concrete
// implementations are constructed and injected. Order: services ->
// repositories -> txManager -> usecases -> handlers.
package bootstrap

import (
	"smart-grid-asset-registry/internal/config"
	"smart-grid-asset-registry/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SetupDependencies wires the application. No feature modules are wired yet;
// the csv-import modules are added by their own tasks.
func SetupDependencies(app *fiber.App, db *gorm.DB, cfg *config.Config, _ *logrus.Logger) *routes.RouteDependencies {
	// --- external services (internal/service/*) ---
	// --- repositories ---
	// --- txManager ---
	// --- usecases ---
	// --- handlers ---

	return &routes.RouteDependencies{
		App: app,
		DB:  db,
		Cfg: cfg,
	}
}
