package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jarlene/tiaozao/internal/handler"
	"github.com/Jarlene/tiaozao/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Setup(
	uploadDir string,
	categoryHandler *handler.CategoryHandler,
	productHandler *handler.ProductHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-Id"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Serve uploaded images
	fileServer := http.FileServer(http.Dir(uploadDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth stub for all API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthStub)

		// Category routes
		r.Route("/categories", func(r chi.Router) {
			r.Post("/", categoryHandler.Create)
			r.Put("/{id}", categoryHandler.Update)
			r.Delete("/{id}", categoryHandler.Delete)
			r.Get("/{id}", categoryHandler.Get)
			r.Get("/", categoryHandler.List)
		})

		// Product routes
		r.Route("/products", func(r chi.Router) {
			r.Post("/", productHandler.Create)
			r.Put("/{id}", productHandler.Update)
			r.Put("/{id}/delist", productHandler.Delist)
			r.Get("/{id}", productHandler.Get)
			r.Get("/", productHandler.List)
			r.Post("/images", productHandler.UploadImage)
		})
	})

	// Serve static frontend for any non-API route
	staticDir := "web"
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		fullPath := filepath.Join(staticDir, path)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, fullPath)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	return r
}
