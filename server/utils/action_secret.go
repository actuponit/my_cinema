package utils

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ActionSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if subtle.ConstantTimeCompare([]byte(c.GetHeader("x-action-secret")), []byte(secret)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Invalid action secret"})
			return
		}
		c.Next()
	}
}
