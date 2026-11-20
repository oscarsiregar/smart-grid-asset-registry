package bootstrap

import (
	"smart-grid-asset-registry/internal/config"
	"smart-grid-asset-registry/internal/middleware"
	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func FiberConfig(cfg *config.Config, log *logrus.Logger) fiber.Config {
	// Behind the reverse proxies of deploy/, c.IP() (and so the limiter)
	// must see the visitor, not the proxy: TRUSTED_PROXIES.
	return middleware.WithTrustedProxies(fiber.Config{
		// Keep Prefork OFF on Kubernetes (FIBER_PREFORK=false): the K8s
		// Service already load-balances and prefork fights CPU limits.
		Prefork:       cfg.FiberPrefork,
		CaseSensitive: true,
		AppName:       cfg.AppName,
		// Global error handler: maps apperror.AppError -> {code,status,message}
		// with the right HTTP status, logs >= 500, hides unknown errors as 500.
		ErrorHandler:   utils.NewErrorHandler(log),
		JSONEncoder:    sonic.Marshal,
		JSONDecoder:    sonic.Unmarshal,
		ReadBufferSize: 65536,
		// 6 MB: a 5 MB CSV plus multipart overhead must reach the usecase,
		// which enforces the 5 MB file limit with its own error code.
		BodyLimit: 6 * 1024 * 1024,
	}, cfg.TrustedProxies)
}
