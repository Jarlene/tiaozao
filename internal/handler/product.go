package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Jarlene/tiaozao/internal/middleware"
	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/service"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Price         float64 `json:"price"`
		OriginalPrice float64 `json:"original_price"`
		CategoryID    *uint   `json:"category_id"`
		ImageIDs      []uint  `json:"image_ids"`
		CoverImageID  *uint   `json:"cover_image_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	if req.Price < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "price must be non-negative"})
		return
	}

	sellerID, _ := strconv.ParseUint(userID, 10, 64)
	var origPrice *float64
	if req.OriginalPrice > 0 {
		origPrice = &req.OriginalPrice
	}

	product := &model.Product{
		Title:         req.Title,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: origPrice,
		Status:        model.ProductStatusDraft,
		CategoryID:    req.CategoryID,
		SellerID:      uint(sellerID),
	}

	if err := h.svc.Create(r.Context(), product, req.ImageIDs, req.CoverImageID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Title         *string  `json:"title"`
		Description   *string  `json:"description"`
		Price         *float64 `json:"price"`
		OriginalPrice *float64 `json:"original_price"`
		CategoryID    *uint    `json:"category_id"`
		Status        *string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	updates := map[string]any{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		updates["original_price"] = *req.OriginalPrice
	}
	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.svc.Update(r.Context(), uint(id), userID, updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *ProductHandler) Delist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	userID := middleware.GetUserID(r.Context())
	if err := h.svc.Delist(r.Context(), uint(id), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "delisted"})
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	product, err := h.svc.GetByID(r.Context(), uint(id))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	status := r.URL.Query().Get("status")

	var categoryID *uint
	if cid := r.URL.Query().Get("category_id"); cid != "" {
		if id, err := strconv.ParseUint(cid, 10, 64); err == nil {
			uid := uint(id)
			categoryID = &uid
		}
	}

	var sellerID *uint
	if sid := r.URL.Query().Get("seller_id"); sid != "" {
		if id, err := strconv.ParseUint(sid, 10, 64); err == nil {
			uid := uint(id)
			sellerID = &uid
		}
	}

	products, total, err := h.svc.List(r.Context(), page, size, categoryID, status, sellerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total": total,
		"items": products,
		"page":  page,
		"size":  size,
	})
}

func (h *ProductHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported file type: " + ext})
		return
	}

	objectName := fmt.Sprintf("%d/%s%s", time.Now().UnixNano(), randomString(8), ext)
	url, err := h.svc.SaveImage(r.Context(), objectName, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Create image record
	imgID, err := h.svc.CreateImageRecord(r.Context(), url)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": imgID, "url": url})
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
