package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Jarlene/tiaozao/internal/config"
	"github.com/Jarlene/tiaozao/internal/handler"
	"github.com/Jarlene/tiaozao/internal/repository"
	"github.com/Jarlene/tiaozao/internal/router"
	"github.com/Jarlene/tiaozao/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Database connection
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Repositories
	categoryRepo := repository.NewCategoryRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)

	// Services
	imageSvc := service.NewImageService(cfg.ImageDir, cfg.ImageBaseURL)
	categorySvc := service.NewCategoryService(categoryRepo)
	productSvc := service.NewProductService(productRepo, imageSvc)
	orderSvc := service.NewOrderService(orderRepo)

	// Handlers
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	productHandler := handler.NewProductHandler(productSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)

	// Router
	r := router.Setup(cfg.ImageDir, categoryHandler, productHandler, orderHandler)

	addr := ":" + cfg.ServerPort
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
