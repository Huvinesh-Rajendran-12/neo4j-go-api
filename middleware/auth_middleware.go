package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/Huvinesh-Rajendran-12/neo4j-go-api/utils"
)

// AuthenticationMiddleware checks if the user has a valid JWT token
func AuthenticationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
            c.Abort()
            return
        }

        // The token should be prefixed with "Bearer "
        tokenParts := strings.Split(tokenString, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
            c.Abort()
            return
        }

        tokenString = tokenParts[1]

        claims, err := utils.VerifyToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
            c.Abort()
            return
        }

        c.Set("secret_id", claims["secret_id"])
        c.Next()
    }
}
