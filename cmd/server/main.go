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
	"ainyx-user-api/db/sqlc"
	"ainyx-user-api/internal/handler"
	"ainyx-user-api/internal/logger"
	"ainyx-user-api/internal/middleware"
	"ainyx-user-api/internal/repository"
	"ainyx-user-api/internal/routes"
	"ainyx-user-api/internal/service"
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

	// Connection pool: safe for concurrent use across requests.
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

	// Wire the layers: sqlc queries -> repository -> service -> handler.
	queries := sqlc.New(pool)
	userRepo := repository.New(queries)
	userSvc := service.New(userRepo, log)
	userHandler := handler.NewUserHandler(userSvc, log)

	app := fiber.New()

	// Order matters: RequestID runs first so the logger can read the id.
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))

	routes.Register(app, userHandler)

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
