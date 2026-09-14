package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jarlene/tiaozao/internal/config"
	"github.com/Jarlene/tiaozao/internal/handler"
	"github.com/Jarlene/tiaozao/internal/repository"
	"github.com/Jarlene/tiaozao/internal/router"
	"github.com/Jarlene/tiaozao/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Database connection pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("database connected")

	// Repositories (constructor injection; no global database singletons)
	categoryRepo := repository.NewCategoryRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)

	// Services
	imageService := service.NewImageService(cfg.ImageDir, cfg.ImageBaseURL)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, commentRepo, imageService)
	commentService := service.NewCommentService(commentRepo)

	// Handlers
	handlers := router.Handlers{
		Category: handler.NewCategoryHandler(categoryService),
		Product:  handler.NewProductHandler(productService),
		Comment:  handler.NewCommentHandler(commentService),
	}

	r := router.Setup(cfg.ImageDir, cfg.CORSAllowedOrigins, handlers)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down gracefully...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
		cancel()
	}()

	log.Printf("server starting on :%s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("server stopped")
}
