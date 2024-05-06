package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/Huvinesh-Rajendran-12/neo4j-go-api/utils"
)

// Function for logging in
func CreateAPIToken(c *gin.Context) {
    var user struct {
        SecretID string `json:"secret_id"`
        Secret string `json:"secret"`
    }

    // Check user credentials and generate a JWT token
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
        return
    }

// Check if credentials are valid (replace this logic with real authentication)
    // Generate a JWT token
    data, err := utils.GenerateToken(user.SecretID, user.Secret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"authentication":data})
}

func CheckAPITokenExpirations(c *gin.Context) {
    var token struct {
        Token string `json:"token"`
    }
    // Check user credentials and generate a JWT token
    if err := c.ShouldBindJSON(&token); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
        return
    }
    data := utils.IsTokenExpired(token.Token)
    c.JSON(http.StatusOK, gin.H{"validity": data})
}
