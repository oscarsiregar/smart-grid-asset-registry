package bootstrap

import (
	"backend-template/internal/config"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/utils"
)

func FiberConfig(cfg *config.Config, log *logrus.Logger) fiber.Config {
	return fiber.Config{
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
		BodyLimit:      4 * 1024 * 1024,
	}
}
