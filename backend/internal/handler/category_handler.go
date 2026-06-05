package handler

import (
	"strconv"

	"flea-market/internal/service"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List 获取分类树
func (h *CategoryHandler) List(c *gin.Context) {
	tree, code, err := h.categoryService.GetTree()
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, tree)
}

// ListFlat 获取扁平分类列表
func (h *CategoryHandler) ListFlat(c *gin.Context) {
	categories, code, err := h.categoryService.GetFlatList()
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, categories)
}

// Create 创建分类
func (h *CategoryHandler) Create(c *gin.Context) {
	var req service.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	category, code, err := h.categoryService.Create(&req)
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

	Success(c, category)
}

// Update 更新分类
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	var req service.UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	category, code, err := h.categoryService.Update(uint(id), &req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 404, code)
		return
	}

	Success(c, category)
}

// Delete 删除分类
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.categoryService.Delete(uint(id))
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

	Success(c, nil)
}
