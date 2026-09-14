package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jarlene/tiaozao/internal/handler"
)

// TestRouterSmoke verifies HTTP router wiring (CORS, health, static serving,
// route registration) without a live database.
func TestRouterSmoke(t *testing.T) {
	h := Handlers{
		Category: &handler.CategoryHandler{},
		Product:  &handler.ProductHandler{},
		Comment:  &handler.CommentHandler{},
	}
	r := Setup("./does-not-exist", []string{"http://localhost:8080"}, h)

	// Health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health: got %d, want 200", w.Code)
	}

	// CORS preflight reflects allowed origin
	req = httptest.NewRequest(http.MethodOptions, "/api/products", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8080" {
		t.Errorf("CORS allow-origin = %q, want http://localhost:8080", got)
	}

	// Disallowed origin not echoed as allowed
	req = httptest.NewRequest(http.MethodOptions, "/api/products", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got == "http://evil.example.com" {
		t.Errorf("CORS should not echo disallowed origin, got %q", got)
	}

	// API routes are registered: a GET to a product detail route must NOT
	// fall through to the static-file 404 handler. With nil services the
	// handler panics and the Recoverer middleware returns 500, proving the
	// route was matched.
	req = httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Errorf("product detail route should be registered, got 404")
	}
}
