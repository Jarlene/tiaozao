package handler

import (
	"net/http"
	"strconv"

	"github.com/Jarlene/tiaozao/backend/internal/middleware"
	"github.com/Jarlene/tiaozao/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// OrderHandler 订单HTTP处理器
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Create 创建订单
// POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	buyerID := middleware.GetUserID(c)

	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	resp, err := h.orderService.CreateOrder(buyerID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "下单成功",
		"data":    resp,
	})
}

// GetByID 获取订单详情
// GET /api/v1/orders/:id
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的订单ID"})
		return
	}

	order, err := h.orderService.GetOrderByID(uint(orderID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    order,
	})
}

// ListBuyerOrders 买家订单列表
// GET /api/v1/orders/buyer
func (h *OrderHandler) ListBuyerOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), 20)
	status := c.Query("status")

	orders, total, err := h.orderService.GetBuyerOrders(userID, page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"items": orders,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// ListSellerOrders 卖家订单列表
// GET /api/v1/orders/seller
func (h *OrderHandler) ListSellerOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), 20)
	status := c.Query("status")

	orders, total, err := h.orderService.GetSellerOrders(userID, page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"items": orders,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// UpdateStatus 更新订单状态
// PUT /api/v1/orders/:id/status
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的订单ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(uint(orderID), userID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "状态更新成功",
		"data":    order,
	})
}
