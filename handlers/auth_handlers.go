package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/Huvinesh-Rajendran-12/neo4j-go-api/utils"
)

// Function for logging in
func CreateAPIToken(c *gin.Context) {
    var user struct {ID int `json:"id"`}

    // Check user credentials and generate a JWT token
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
        return
    }

// Check if credentials are valid (replace this logic with real authentication)
    // Generate a JWT token
    token, err := utils.GenerateToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}

