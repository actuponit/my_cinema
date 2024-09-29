package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}
		authSlice := strings.Split(header, " ")
		if len(authSlice) != 2 || strings.ToLower(authSlice[0]) != "bearer" {
			c.JSON(401, gin.H{"message": "Invalid authorization header"})
			c.Abort()
			return
		}
		claims, err := ValidateToken(authSlice[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		hasuraClaims, ok := claims["https://hasura.io/jwt/claims"].(map[string]interface{})
		if ok {
			c.Set("id", hasuraClaims["x-hasura-user-id"].(string))
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		userRole := hasuraClaims["x-hasura-default-role"].(string)
		for _, role := range roles {
			if role == userRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		c.Abort()
	}
}
