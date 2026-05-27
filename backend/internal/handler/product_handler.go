package handler

import (
	"net/http"
	"strconv"

	"github.com/Jarlene/tiaozao/backend/internal/middleware"
	"github.com/Jarlene/tiaozao/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ProductHandler 商品HTTP处理器
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler 创建商品处理器实例
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// Create 创建商品
// POST /api/v1/products
func (h *ProductHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	product, err := h.productService.Create(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    product,
	})
}

// GetByID 获取商品详情
// GET /api/v1/products/:id
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    product,
	})
}

// Update 更新商品
// PUT /api/v1/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	product, err := h.productService.Update(userID, uint(productID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    product,
	})
}

// Delete 删除商品
// DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	if err := h.productService.Delete(userID, uint(productID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// List 商品列表
// GET /api/v1/products
func (h *ProductHandler) List(c *gin.Context) {
	var query service.ProductListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "查询参数错误", "error": err.Error()})
		return
	}

	products, total, err := h.productService.List(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"items": products,
			"total": total,
			"page":  query.Page,
			"size":  query.PageSize,
		},
	})
}

// Search 搜索商品
// GET /api/v1/products/search
func (h *ProductHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入搜索关键词"})
		return
	}

	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), 20)

	query := &service.ProductListQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
	}

	products, total, err := h.productService.List(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"items":   products,
			"total":   total,
			"page":    page,
			"size":    pageSize,
			"keyword": keyword,
		},
	})
}

// UploadImage 上传图片
// POST /api/v1/products/upload
func (h *ProductHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择要上传的图片"})
		return
	}

	// 保存文件到本地uploads目录
	filePath := "uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "图片上传失败"})
		return
	}

	imageURL, err := h.productService.UploadImage(file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
		"data": gin.H{
			"url": imageURL,
		},
	})
}

// parseInt 辅助函数：将字符串转为整数，转换失败时返回默认值
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
