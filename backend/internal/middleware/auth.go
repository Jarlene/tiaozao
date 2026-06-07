package middleware

import (
	"net/http"
	"strings"

	"flea-market/internal/config"
	"flea-market/internal/handler"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.Error(c, http.StatusUnauthorized, errors.ErrInvalidToken)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			handler.Error(c, http.StatusUnauthorized, errors.ErrInvalidToken)
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})
		if err != nil || !token.Valid {
			handler.Error(c, http.StatusUnauthorized, errors.ErrInvalidToken)
			c.Abort()
			return
		}

		// 验证是 access token 而非 refresh token
		if claims.Subject != "access" {
			handler.Error(c, http.StatusUnauthorized, errors.ErrInvalidToken)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
