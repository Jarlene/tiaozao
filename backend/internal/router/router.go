package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jarlene/tiaozao/internal/handler"
	"github.com/jarlene/tiaozao/internal/middleware"
)

func Setup(authHandler *handler.AuthHandler, userHandler *handler.UserHandler) *gin.Engine {
	r := gin.Default()

	// Global middleware
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		// Public auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/send-code", authHandler.SendCode)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		user := api.Group("/user")
		user.Use(middleware.AuthRequired())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
		}
	}

	return r
}
