package main

import (
	"context"
	"fmt"
	"time"

	"smart-grid-asset-registry/internal/bootstrap"
	"smart-grid-asset-registry/internal/config"
	"smart-grid-asset-registry/internal/middleware"
	"smart-grid-asset-registry/internal/routes"
	"smart-grid-asset-registry/pkg/gocommon/pkg/database"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
	commonMiddleware "smart-grid-asset-registry/pkg/gocommon/pkg/middleware/fiber"
	"smart-grid-asset-registry/pkg/gocommon/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/sirupsen/logrus"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	log := logger.New(cfg.IsProd)

	// database.Connect already registers go-common's audit plugin — do NOT
	// call db.Use(database.NewAuditPlugin()) again (gorm.ErrRegistered).
	db, err := database.Connect(&database.DBConfig{
		Host:            cfg.DBHost,
		User:            cfg.DBUser,
		Password:        cfg.DBPassword,
		DBName:          cfg.DBName,
		Port:            cfg.DBPort,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	}, log, cfg.IsProd)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if !cfg.IsProd {
		db.Logger = gormlogger.Default.LogMode(gormlogger.Info)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	app := fiber.New(bootstrap.FiberConfig(cfg, log))

	// Middleware order matters: recover first so it wraps everything.
	app.Use(middleware.RecoverConfig())
	commonMiddleware.InstallCommonMiddleware(app, log) // request id, logging, security headers, telemetry
	app.Use(middleware.LimiterConfig())
	app.Use(compress.New())

	routeDeps := bootstrap.SetupDependencies(app, db, cfg, log)
	routes.RegisterRoutes(routeDeps)

	app.Use(utils.NotFoundHandler)

	address := fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort)

	go func() {
		log.Infof("Server starting on %s", address)
		if err := app.Listen(address); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	sh := utils.NewShutdownHelper(log)
	sh.Wait()

	log.Info("Shutting down Fiber Server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Errorf("Fiber Server shutdown error: %v", err)
	} else {
		log.Info("Fiber Server shut down successfully")
	}
	cancel()

	sh.Graceful(map[string]func(ctx context.Context) error{
		"Database": func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	log.Info("Server exited")
}
