package router

import (
	"flea-market/internal/chat"
	"flea-market/internal/config"
	"flea-market/internal/handler"
	"flea-market/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Setup(cfg *config.Config, authHandler *handler.AuthHandler, productHandler *handler.ProductHandler, categoryHandler *handler.CategoryHandler, reviewHandler *handler.ReviewHandler, chatHandler *handler.ChatHandler, wsHandler *chat.WSHandler, orderHandler *handler.OrderHandler, logger *zap.Logger) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 认证路由（无需登录）
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// 受保护路由
	protected := r.Group("/api/v1/auth")
	protected.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)
	}

	// 公开商品路由（无需登录）
	publicProducts := r.Group("/api/v1/products")
	{
		publicProducts.GET("", productHandler.List)
		publicProducts.GET("/search", productHandler.Search)
		publicProducts.GET("/:id", productHandler.GetByID)
	}

	// 受保护商品路由
	protectedProducts := r.Group("/api/v1/products")
	protectedProducts.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		// 注意：/mine 必须在 /:id 之前注册，避免 Gin 路由冲突
		protectedProducts.GET("/mine", productHandler.ListMine)
		protectedProducts.GET("/mine/counts", productHandler.GetCounts)
		protectedProducts.POST("", productHandler.Create)
		protectedProducts.PUT("/:id", productHandler.Update)
		protectedProducts.DELETE("/:id", productHandler.Delete)
		protectedProducts.PUT("/:id/status", productHandler.UpdateStatus)
	}

	// 图片路由
	protectedImages := r.Group("/api/v1/images")
	protectedImages.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		protectedImages.POST("/upload", productHandler.UploadImage)
		protectedImages.GET("/:id", productHandler.GetImage)
		protectedImages.DELETE("/:id", productHandler.DeleteImage)
	}

	// 分类路由（列表公开，管理需认证）
	r.GET("/api/v1/categories", categoryHandler.List)
	r.GET("/api/v1/categories/flat", categoryHandler.ListFlat)

	protectedCategories := r.Group("/api/v1/categories")
	protectedCategories.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		protectedCategories.POST("", categoryHandler.Create)
		protectedCategories.PUT("/:id", categoryHandler.Update)
		protectedCategories.DELETE("/:id", categoryHandler.Delete)
	}

	// 评论路由（公开：列表和统计）
	publicReviews := r.Group("/api/v1/products/:id/reviews")
	{
		publicReviews.GET("", reviewHandler.List)
		publicReviews.GET("/stats", reviewHandler.GetStats)
	}

	// 评论路由（需要登录：发布和检查）
	protectedReviews := r.Group("/api/v1/products/:id/reviews")
	protectedReviews.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		protectedReviews.POST("", reviewHandler.Create)
		protectedReviews.GET("/check", reviewHandler.CheckCanReview)
	}

	// 回复路由（需要登录）
	protectedReplies := r.Group("/api/v1")
	protectedReplies.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		protectedReplies.POST("/reviews/:reviewId/reply", reviewHandler.Reply)
	}

	// WebSocket 路由（通过 token 参数认证）
	r.GET("/api/v1/ws", wsHandler.HandleWebSocket)

	// 聊天路由（需要登录）
	protectedChat := r.Group("/api/v1")
	protectedChat.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		protectedChat.GET("/conversations", chatHandler.ListConversations)
		protectedChat.POST("/conversations", chatHandler.CreateConversation)
		protectedChat.GET("/conversations/:id/messages", chatHandler.GetMessages)
		protectedChat.POST("/conversations/:id/read", chatHandler.MarkRead)
		protectedChat.POST("/messages", chatHandler.SendMessage)
	}

	// 订单路由（需要登录）
	protectedOrders := r.Group("/api/v1/orders")
	protectedOrders.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		// 注意：/mine 和 /sold 必须在 /:id 之前注册
		protectedOrders.GET("/mine", orderHandler.ListMine)
		protectedOrders.GET("/sold", orderHandler.ListSold)
		protectedOrders.POST("", orderHandler.Create)
		protectedOrders.GET("/:id", orderHandler.GetByID)
		protectedOrders.POST("/:id/cancel", orderHandler.Cancel)
		protectedOrders.POST("/:id/pay", orderHandler.Pay)
		protectedOrders.POST("/:id/ship", orderHandler.Ship)
		protectedOrders.POST("/:id/confirm", orderHandler.ConfirmReceive)
		protectedOrders.POST("/:id/refund", orderHandler.RequestRefund)
		protectedOrders.POST("/:id/refund/complete", orderHandler.CompleteRefund)
		protectedOrders.POST("/:id/dispute", orderHandler.RaiseDispute)
		protectedOrders.POST("/:id/arbitrate", orderHandler.Arbitrate)
		protectedOrders.GET("/:id/logs", orderHandler.GetStatusLogs)
	}

	return r
}
