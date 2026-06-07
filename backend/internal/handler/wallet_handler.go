package handler

import (
	"strconv"

	"flea-market/internal/service"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// GetWallet 获取钱包信息
func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	info, code, err := h.walletService.GetWalletInfo(uint(userID))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, info)
}

// TopUp 充值
func (h *WalletHandler) TopUp(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	var req service.TopUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	info, code, err := h.walletService.TopUp(uint(userID), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, info)
}

// ListTransactions 获取交易流水
func (h *WalletHandler) ListTransactions(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	items, total, code, err := h.walletService.ListTransactions(uint(userID), page, pageSize)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, gin.H{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}
