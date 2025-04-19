package routes

import (
	"log"
	"net/http"
	"strings"

	"lostfound-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Validate the token
		token, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract the claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Debug: Print the claims to log for verification
		log.Println("Token Claims:", claims)

		// Check if "userId" exists in the claims and extract it
		userID, exists := claims["userId"].(string)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing userId in token"})
			c.Abort()
			return
		}

		// Store the userId in the context for further use in the route handlers
		c.Set("userID", userID)

		// Proceed with the next middleware/handler
		c.Next()
	}
}
