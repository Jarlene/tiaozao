package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthStub is a development auth stub that extracts user info from the
// X-User-Id header or injects a default test user.
// Replace with real JWT middleware when JAR-7 (auth system) is implemented.
func AuthStub(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-Id")
		if userID == "" {
			userID = "1" // Default test user for development
		}

		// Simple token validation from Authorization header (stub)
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != "dev-token" && !strings.HasPrefix(token, "eyJ") {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}
