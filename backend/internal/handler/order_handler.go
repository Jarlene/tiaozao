package handler

import (
	"strconv"

	"flea-market/internal/model"
	"flea-market/internal/service"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Create 创建订单
func (h *OrderHandler) Create(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	var req service.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	order, code, err := h.orderService.CreateOrder(uint(userID), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, order)
}

// Cancel 取消订单
func (h *OrderHandler) Cancel(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.Cancel(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// Pay 付款
func (h *OrderHandler) Pay(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.Pay(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// Ship 卖家发货
func (h *OrderHandler) Ship(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.ShipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.Ship(uint(userID), uint(id), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// ConfirmReceive 确认收货
func (h *OrderHandler) ConfirmReceive(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.ConfirmReceive(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// RequestRefund 申请退款
func (h *OrderHandler) RequestRefund(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.RequestRefundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.RequestRefund(uint(userID), uint(id), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// ApproveRefund 卖家同意退款
func (h *OrderHandler) ApproveRefund(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.ApproveRefund(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// RejectRefund 卖家拒绝退款
func (h *OrderHandler) RejectRefund(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.RejectRefund(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// CompleteRefund 退款成功
func (h *OrderHandler) CompleteRefund(c *gin.Context) {
	// 管理员/系统操作 — 当前复用 user_id 作为操作者 ID
	operatorID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.CompleteRefund(uint(operatorID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// RaiseDispute 发起纠纷
func (h *OrderHandler) RaiseDispute(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.RaiseDispute(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// Arbitrate 管理员仲裁
func (h *OrderHandler) Arbitrate(c *gin.Context) {
	adminID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.ArbitrateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.orderService.Arbitrate(uint(adminID), uint(id), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, nil)
}

// GetByID 获取订单详情
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	order, code, err := h.orderService.GetByID(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, order)
}

// ListMine 我的订单列表（作为买家）
func (h *OrderHandler) ListMine(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var status *model.OrderStatus
	if s := c.Query("status"); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			st := model.OrderStatus(val)
			status = &st
		}
	}

	result, code, err := h.orderService.ListByBuyer(uint(userID), status, page, pageSize)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, result)
}

// ListSold 我的订单列表（作为卖家）
func (h *OrderHandler) ListSold(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var status *model.OrderStatus
	if s := c.Query("status"); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			st := model.OrderStatus(val)
			status = &st
		}
	}

	result, code, err := h.orderService.ListBySeller(uint(userID), status, page, pageSize)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, result)
}

// GetStatusLogs 获取订单状态变更日志
func (h *OrderHandler) GetStatusLogs(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	logs, code, err := h.orderService.GetStatusLogs(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if orderErrorResponse(c, code) {
		return
	}

	Success(c, logs)
}

// ListDisputes 管理员获取纠纷订单列表
func (h *OrderHandler) ListDisputes(c *gin.Context) {
	adminID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, code, err := h.orderService.ListDisputes(uint(adminID), page, pageSize)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		orderErrorResponse(c, code)
		return
	}

	Success(c, result)
}

// orderErrorResponse 将业务错误码转为 HTTP 状态码并返回错误响应
// 返回 true 表示已处理（调用方应 return），false 表示无需处理
func orderErrorResponse(c *gin.Context, code int) bool {
	if code == errors.Success {
		return false
	}
	httpStatus := 400
	switch code {
	case errors.ErrProductNotFound, errors.ErrOrderNotFound:
		httpStatus = 404
	case errors.ErrForbidden, errors.ErrForbiddenNotAdmin:
		httpStatus = 403
	case errors.ErrInsufficientBalance:
		httpStatus = 422 // 余额不足
	case errors.ErrInsufficientStock:
		httpStatus = 409 // 库存不足（冲突）
	}
	Error(c, httpStatus, code)
	return true
}
