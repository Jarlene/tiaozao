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

type Handlers struct {
	Category *handler.CategoryHandler
	Product  *handler.ProductHandler
	Comment  *handler.CommentHandler
}

func Setup(
	uploadDir string,
	allowedOrigins []string,
	h Handlers,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:8080"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
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
			r.Post("/", h.Category.Create)
			r.Put("/{id}", h.Category.Update)
			r.Delete("/{id}", h.Category.Delete)
			r.Get("/{id}", h.Category.Get)
			r.Get("/", h.Category.List)
		})

		// Product routes
		r.Route("/products", func(r chi.Router) {
			r.Post("/", h.Product.Create)
			r.Put("/{id}", h.Product.Update)
			r.Put("/{id}/delist", h.Product.Delist)
			r.Get("/{id}", h.Product.Get)
			r.Get("/", h.Product.List)
			r.Post("/images", h.Product.UploadImage)

			// Comment routes (nested under products)
			r.Route("/{productId}/comments", func(r chi.Router) {
				r.Get("/", h.Comment.ListComments)
				r.Post("/", h.Comment.CreateComment)
				r.Post("/{commentId}/reply", h.Comment.ReplyToComment)
			})
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
