package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jarlene/tiaozao/internal/middleware"
	"github.com/Jarlene/tiaozao/internal/service"
	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

// ListComments GET /api/v1/products/{productId}/comments
func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseUint(chi.URLParam(r, "productId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的商品ID"})
		return
	}

	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("page_size"), 10)

	result, err := h.svc.ListComments(r.Context(), uint(productID), page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateComment POST /api/v1/products/{productId}/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseUint(chi.URLParam(r, "productId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的商品ID"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "内容不能为空"})
		return
	}

	comment, err := h.svc.CreateComment(r.Context(), uint(productID), userID, strings.TrimSpace(req.Content))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

// ReplyToComment POST /api/v1/products/{productId}/comments/{commentId}/reply
func (h *CommentHandler) ReplyToComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.ParseUint(chi.URLParam(r, "commentId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的评论ID"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "回复内容不能为空"})
		return
	}

	comment, err := h.svc.ReplyToComment(r.Context(), uint(commentID), userID, strings.TrimSpace(req.Content))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultVal
	}
	return v
}
