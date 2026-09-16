package middleware

import (
	"net/http"
	"strings"

	"agnos-assessment/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyStaffID  = "staff_id"
	ContextKeyUsername = "username"
	ContextKeyHospital = "hospital"
)

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization token is missing",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization header format must be 'Bearer <token>'",
			})
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := authService.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or expired authorization token",
			})
			return
		}

		// Store verified staff identity in Gin context
		c.Set(ContextKeyStaffID, claims.StaffID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyHospital, claims.Hospital)

		c.Next()
	}
}
