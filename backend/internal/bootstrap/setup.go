// Package bootstrap is the composition root: the ONLY place concrete
// implementations are constructed and injected. Order: services ->
// repositories -> txManager -> usecases -> handlers.
package bootstrap

import (
	"backend-template/internal/config"
	"backend-template/internal/handler/product"
	"backend-template/internal/repository/psql/gormtx"
	product_repository "backend-template/internal/repository/psql/product"
	"backend-template/internal/routes"
	product_usecase "backend-template/internal/usecase/product"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SetupDependencies(app *fiber.App, db *gorm.DB, cfg *config.Config, log *logrus.Logger) *routes.RouteDependencies {
	// --- external services (internal/service/*) ---
	// e.g. userManagementService := user_management_service.NewUserManagementService(cfg.UserManagementBaseURL, cfg.InternalServiceSecret, log)

	// --- repositories ---
	productRepository := product_repository.NewProductRepository(db)

	txManager := gormtx.NewTransactionManager(db)

	// --- usecases ---
	productUsecase := product_usecase.NewProductUsecase(txManager, productRepository, log)

	// --- handlers ---
	productHandler := product.NewProductHandler(productUsecase)

	return &routes.RouteDependencies{
		App:            app,
		DB:             db,
		Cfg:            cfg,
		ProductHandler: productHandler,
	}
}
