package server

import (
	"ccards/internal/card"
	"ccards/internal/client"
	"ccards/internal/router"
	"ccards/pkg/config"
	"ccards/pkg/database"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Bootstrap struct {
	config *config.Config
	db     *sql.DB
	router *gin.Engine
	redis  *redis.Client
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

func (b *Bootstrap) Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	b.config = cfg

	log.Printf("Starting %s in %s environment", cfg.App.Name, cfg.App.Environment)

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	b.db = db

	if err := b.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	b.redis = redisClient

	// client
	clientRepo := client.NewRepository(db)
	clientService := client.NewService(clientRepo, cfg.JWT, b.redis)
	clientHandler := client.NewHandler(clientService)

	// cards
	cardRepo := card.NewRepository(db)
	cardService := card.NewService(cardRepo)
	cardHandler := card.NewHandler(cardService)

	r := router.NewRouter(router.RouterConfig{
		ClientHandler: clientHandler,
		CardHandler:   cardHandler,
		Config:        b.config,
		RedisClient:   b.redis,
	})

	b.router = r.Setup()

	srv := &http.Server{
		Addr:         b.config.Server.GetAddress(),
		Handler:      b.router,
		ReadTimeout:  b.config.Server.ReadTimeout,
		WriteTimeout: b.config.Server.WriteTimeout,
		IdleTimeout:  b.config.Server.IdleTimeout,
	}

	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := b.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited")
	return nil
}

func (b *Bootstrap) runMigrations() error {
	log.Println("Running database migrations...")

	if b.config.App.Environment == "local" {
		if err := database.RunMigrations(b.db, "./db/migrations"); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func (b *Bootstrap) GetDB() *sql.DB {
	return b.db
}

func (b *Bootstrap) GetRouter() *gin.Engine {
	return b.router
}

func (b *Bootstrap) GetConfig() *config.Config {
	return b.config
}
