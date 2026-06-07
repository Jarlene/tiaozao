package handler

import (
	"strconv"

	"flea-market/internal/model"
	"flea-market/internal/service"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// Create 创建商品
func (h *ProductHandler) Create(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	var req service.CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	// 验证图片数量
	if len(req.ImageIDs) > 9 {
		Error(c, 400, errors.ErrTooManyFiles)
		return
	}

	product, code, err := h.productService.Create(uint(userID), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrCategoryNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, product)
}

// GetByID 获取商品详情
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	product, code, err := h.productService.GetByID(uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 404, code)
		return
	}

	Success(c, product)
}

// Update 更新商品
func (h *ProductHandler) Update(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	product, code, err := h.productService.Update(uint(userID), uint(id), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 403
		if code == errors.ErrProductNotFound {
			httpStatus = 404
		} else if code == errors.ErrCategoryNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, product)
}

// Delete 下架商品
func (h *ProductHandler) Delete(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.productService.Delete(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 403
		if code == errors.ErrProductNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, nil)
}

// List 商品列表（公开，在售）
func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, code, err := h.productService.List(page, pageSize)
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

// Search 搜索商品
func (h *ProductHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keyword := c.Query("keyword")

	var categoryID *uint
	if cid := c.Query("category_id"); cid != "" {
		if id, err := strconv.ParseUint(cid, 10, 32); err == nil {
			categoryID = new(uint)
			*categoryID = uint(id)
		}
	}

	priceMin, _ := strconv.ParseInt(c.Query("price_min"), 10, 64)
	priceMax, _ := strconv.ParseInt(c.Query("price_max"), 10, 64)

	result, code, err := h.productService.Search(keyword, categoryID, priceMin, priceMax, page, pageSize)
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

// ListMine 我的商品列表
func (h *ProductHandler) ListMine(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var status *model.ProductStatus
	if s := c.Query("status"); s != "" {
		if val, err := strconv.Atoi(s); err == nil {
			st := model.ProductStatus(val)
			if st == model.ProductStatusInactive || st == model.ProductStatusActive || st == model.ProductStatusSold {
				status = &st
			}
		}
	}

	result, code, err := h.productService.ListMyProducts(uint(userID), status, page, pageSize)
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

// GetCounts 获取我的商品数量统计
func (h *ProductHandler) GetCounts(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	counts, code, err := h.productService.GetMyCounts(uint(userID))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, counts)
}

// UpdateStatus 更新商品状态
func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req struct {
		Status model.ProductStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	// 校验状态值合法性
	if req.Status != model.ProductStatusActive && req.Status != model.ProductStatusInactive && req.Status != model.ProductStatusSold {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.productService.UpdateStatus(uint(userID), uint(id), req.Status)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 403
		if code == errors.ErrProductNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, nil)
}

// UploadImage 上传图片
func (h *ProductHandler) UploadImage(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}
	defer file.Close()

	image, code, err := h.productService.UploadImage(uint(userID), file, header)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, image)
}

// GetImage 获取图片信息
func (h *ProductHandler) GetImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	image, code, err := h.productService.GetImage(uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 404, code)
		return
	}

	Success(c, image)
}

// DeleteImage 删除图片
func (h *ProductHandler) DeleteImage(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.productService.DeleteImage(uint(userID), uint(id))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrImageNotFound {
			httpStatus = 404
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, nil)
}
