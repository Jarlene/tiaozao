package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Jarlene/tiaozao/internal/middleware"
	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/service"
	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create 创建订单
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := parseUint(middleware.GetUserID(r.Context()))

	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求参数"})
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "订单至少需要一个商品"})
		return
	}
	for _, item := range req.Items {
		if item.ProductID == 0 || item.Quantity < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "商品参数无效"})
			return
		}
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

// GetByID 获取订单详情
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := parseUint(middleware.GetUserID(r.Context()))

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的订单ID"})
		return
	}

	order, err := h.svc.GetOrderByID(r.Context(), uint(id), userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, order)
}

// List 获取订单列表
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := parseUint(middleware.GetUserID(r.Context()))

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	status := r.URL.Query().Get("status")
	role := r.URL.Query().Get("role")
	if role != "seller" {
		role = "buyer"
	}

	orders, total, err := h.svc.ListOrders(r.Context(), userID, role, page, size, status)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if orders == nil {
		orders = []model.Order{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": orders,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// Transition 更新订单状态
func (h *OrderHandler) Transition(w http.ResponseWriter, r *http.Request) {
	userID := parseUint(middleware.GetUserID(r.Context()))

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的订单ID"})
		return
	}

	var req model.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求参数"})
		return
	}

	order, err := h.svc.TransitionStatus(r.Context(), uint(id), userID, req.Action)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func parseUint(s string) uint {
	var id uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		id = id*10 + uint(c-'0')
	}
	return id
}
