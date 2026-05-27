package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/service"
	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		ParentID  *uint  `json:"parent_id"`
		SortOrder int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	cat := &model.Category{Name: req.Name, ParentID: req.ParentID, SortOrder: req.SortOrder}
	if err := h.svc.Create(r.Context(), cat); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		Name      *string `json:"name"`
		ParentID  *uint   `json:"parent_id"`
		SortOrder *int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if err := h.svc.Update(r.Context(), uint(id), updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(r.Context(), uint(id)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	cat, err := h.svc.GetByID(r.Context(), uint(id))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cats)
}
