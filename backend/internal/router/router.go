package router

import (
	"github.com/Jarlene/tiaozao/backend/internal/handler"
	"github.com/Jarlene/tiaozao/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Setup 配置所有路由
func Setup(
	r *gin.Engine,
	authMiddleware *middleware.JWTAuth,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
) {
	// 全局中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "tiaozao-backend"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")

	// 认证相关（无需登录）
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// 商品相关（公开接口）
	products := v1.Group("/products")
	{
		products.GET("", productHandler.List)
		products.GET("/search", productHandler.Search)
		products.GET("/:id", productHandler.GetByID)
	}

	// 需要登录的路由组
	api := v1.Group("")
	api.Use(authMiddleware.Middleware())
	{
		// 用户相关
		user := api.Group("/user")
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/products", userHandler.GetUserProducts)
		}

		// 商品管理（需要登录）
		productMgmt := api.Group("/products")
		{
			productMgmt.POST("", productHandler.Create)
			productMgmt.PUT("/:id", productHandler.Update)
			productMgmt.DELETE("/:id", productHandler.Delete)
			productMgmt.POST("/upload", productHandler.UploadImage)
		}

		// 订单相关
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.Create)
			orders.GET("/buyer", orderHandler.ListBuyerOrders)
			orders.GET("/seller", orderHandler.ListSellerOrders)
			orders.GET("/:id", orderHandler.GetByID)
			orders.PUT("/:id/status", orderHandler.UpdateStatus)
		}
	}
}
