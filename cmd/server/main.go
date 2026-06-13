package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"ainyx-user-api/config"
	"ainyx-user-api/internal/logger"
)

func main() {
	// Load .env if present. Ignoring the error is intentional:
	// in Docker/CI the variables come from the real environment.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err) // logger doesn't exist yet, so panic is acceptable here
	}

	log, err := logger.New(cfg.AppEnv)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	// Connection pool: safe for concurrent use across requests,
	// unlike a single connection.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to create db pool", zap.Error(err))
	}
	defer pool.Close()

	// Fail fast at startup if the DB is unreachable.
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping database", zap.Error(err))
	}
	log.Info("connected to database")

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Run the server in a goroutine so main can block on shutdown signals.
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal("server stopped", zap.Error(err))
		}
	}()
	log.Info("server started", zap.String("port", cfg.Port))

	// Graceful shutdown: finish in-flight requests on Ctrl+C / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	if err := app.Shutdown(); err != nil {
		log.Error("error during shutdown", zap.Error(err))
	}
}