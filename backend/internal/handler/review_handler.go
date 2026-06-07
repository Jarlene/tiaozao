package handler

import (
	"strconv"

	"flea-market/internal/service"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// CheckCanReview 检查当前用户是否可以评论
func (h *ReviewHandler) CheckCanReview(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	canReview, code, err := h.reviewService.CheckCanReview(uint(userID), uint(productID))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}

	Success(c, gin.H{
		"can_review": canReview,
		"code":       code,
	})
}

// Create 创建评论
func (h *ReviewHandler) Create(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.CreateReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	review, code, err := h.reviewService.CreateReview(uint(userID), uint(productID), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrProductNotFound || code == errors.ErrReviewNotFound {
			httpStatus = 404
		} else if code == errors.ErrForbidden || code == errors.ErrCannotReviewOwn {
			httpStatus = 403
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, review)
}

// List 获取商品评论列表
func (h *ReviewHandler) List(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, code, err := h.reviewService.ListReviews(uint(productID), page, pageSize)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrProductNotFound || code == errors.ErrReviewNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, result)
}

// GetStats 获取评分统计
func (h *ReviewHandler) GetStats(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	stats, code, err := h.reviewService.GetRatingStats(uint(productID))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrProductNotFound || code == errors.ErrReviewNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, stats)
}

// Reply 卖家回复评论
func (h *ReviewHandler) Reply(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	reviewID, err := strconv.ParseUint(c.Param("reviewId"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.ReplyReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	review, code, err := h.reviewService.ReplyReview(uint(userID), uint(reviewID), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrReviewNotFound || code == errors.ErrProductNotFound {
			httpStatus = 404
		} else if code == errors.ErrForbidden {
			httpStatus = 403
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, review)
}
