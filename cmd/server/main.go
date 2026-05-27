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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database connection
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("database connected")

	// Repositories
	categoryRepo := repository.NewCategoryRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)

	// Services
	imageService := service.NewImageService(cfg.ImageDir, cfg.ImageBaseURL)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, imageService)
	commentService := service.NewCommentService(commentRepo)

	// Handlers
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	commentHandler := handler.NewCommentHandler(commentService)

	// Router
	r := router.Setup(cfg.ImageDir, categoryHandler, productHandler, commentHandler)

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down gracefully...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
		cancel()
	}()

	log.Printf("server starting on port %s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
	log.Println("server stopped")
}
