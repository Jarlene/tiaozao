package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jarlene/tiaozao/internal/model"
	"github.com/jarlene/tiaozao/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.userService.GetProfile(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toProfileResponse(user),
	})
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	user, err := h.userService.UpdateProfile(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toProfileResponse(user),
	})
}
