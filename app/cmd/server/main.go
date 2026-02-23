package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/hamfa/task-manager/internal/handler"
	"github.com/hamfa/task-manager/internal/middleware"
	"github.com/hamfa/task-manager/internal/model"
	"github.com/hamfa/task-manager/internal/repository"
	"github.com/hamfa/task-manager/internal/service"
	"github.com/hamfa/task-manager/pkg/config"
)

const version = "1.0.0"

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ── Connect to PostgreSQL ──────────────────────────────────────
	pgPool, err := pgxpool.New(ctx, cfg.PostgresDSN())
	if err != nil {
		logger.Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pgPool.Close()
	logger.Info("connected to PostgreSQL")

	// ── Connect to MongoDB ─────────────────────────────────────────
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}
	defer mongoClient.Disconnect(ctx)
	logger.Info("connected to MongoDB")

	// ── Connect to Redis ───────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	logger.Info("connected to Redis")

	// ── Initialize Repositories ────────────────────────────────────
	postgresRepo := repository.NewPostgresRepository(pgPool)
	mongoRepo := repository.NewMongoRepository(mongoClient.Database(cfg.MongoDB))
	redisCache := repository.NewRedisCache(redisClient)

	// Initialize database schema
	if err := postgresRepo.InitSchema(ctx); err != nil {
		logger.Fatal("failed to initialize schema", zap.Error(err))
	}
	logger.Info("database schema initialized")

	// ── Initialize Service & Handlers ──────────────────────────────
	taskService := service.NewTaskService(postgresRepo, mongoRepo, redisCache, logger)
	taskHandler := handler.NewTaskHandler(taskService)

	// ── Setup Gin Router ───────────────────────────────────────────
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		services := map[string]string{}

		// Check PostgreSQL
		if err := postgresRepo.Ping(c.Request.Context()); err != nil {
			services["postgresql"] = "unhealthy"
		} else {
			services["postgresql"] = "healthy"
		}

		// Check MongoDB
		if err := mongoRepo.Ping(c.Request.Context()); err != nil {
			services["mongodb"] = "unhealthy"
		} else {
			services["mongodb"] = "healthy"
		}

		// Check Redis
		if err := redisCache.Ping(c.Request.Context()); err != nil {
			services["redis"] = "unhealthy"
		} else {
			services["redis"] = "healthy"
		}

		status := "healthy"
		httpCode := http.StatusOK
		for _, v := range services {
			if v == "unhealthy" {
				status = "degraded"
				httpCode = http.StatusServiceUnavailable
				break
			}
		}

		c.JSON(httpCode, model.HealthResponse{
			Status:   status,
			Version:  version,
			Services: services,
		})
	})

	// API routes
	api := router.Group("/api")
	taskHandler.RegisterRoutes(api)

	// ── Start Server with Graceful Shutdown ─────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited gracefully")
}
