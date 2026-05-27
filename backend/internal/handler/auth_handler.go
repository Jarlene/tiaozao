package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jarlene/tiaozao/internal/middleware"
	"github.com/jarlene/tiaozao/internal/model"
	"github.com/jarlene/tiaozao/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SendCode sends a verification code to the user's phone
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req model.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid phone number format",
		})
		return
	}

	if err := h.authService.SendCode(req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to send verification code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "verification code sent",
	})
}

// Login authenticates user with phone and verification code
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	user, err := h.authService.LoginWithCode(req.Phone, req.Code)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid verification code",
			})
		case service.ErrCodeExpired:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "verification code expired",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "login failed",
			})
		}
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, model.LoginResponse{
		Token: token,
		User:  toProfileResponse(user),
	})
}

func toProfileResponse(user *model.User) model.UserProfileResponse {
	return model.UserProfileResponse{
		ID:        user.ID,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}
}
