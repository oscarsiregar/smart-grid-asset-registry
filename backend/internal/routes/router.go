package routes

import (
	"backend-template/internal/config"
	"backend-template/internal/handler/product"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	pkg_middleware_fiber "repo.smf-indonesia.co.id/package-helper/go-common/pkg/middleware/fiber"
	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/utils"
)

type RouteDependencies struct {
	App            *fiber.App
	DB             *gorm.DB
	Cfg            *config.Config
	ProductHandler product.ProductHandler
}

func RegisterRoutes(deps *RouteDependencies) {
	v1 := deps.App.Group("/v1")

	// Public: liveness/readiness probe (pings the DB).
	utils.RegisterHealthCheck(v1, deps.DB, nil)

	// Every business route MUST hang off this group. The chain order is
	// mandatory (see docs/rules/07-security.md):
	//   1. InternalSecretFromEnv — request must come from the BFF
	//   2. Identity              — reads X-User-* headers into ctx.Locals
	//   3. JWTVerifyStrict       — re-verifies the JWT, 401 when missing/invalid
	// Never register a business route on v1 or deps.App directly.
	authenticated := v1.Group("",
		pkg_middleware_fiber.InternalSecretFromEnv(deps.Cfg.InternalServiceSecret),
		pkg_middleware_fiber.Identity(),
		pkg_middleware_fiber.JWTVerifyStrict(deps.Cfg.JWTSecret),
	)

	productGroup := authenticated.Group("/products")
	productGroup.Get("/", deps.ProductHandler.GetListProduct)
	productGroup.Post("/", deps.ProductHandler.CreateProduct)
	productGroup.Get("/:id", deps.ProductHandler.GetProductDetail)
	productGroup.Put("/:id", deps.ProductHandler.UpdateProduct)
	productGroup.Delete("/:id", deps.ProductHandler.DeleteProduct)
}
