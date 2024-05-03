package main

import (
	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/handlers"
	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	r := gin.Default()
    api := r.Group("/api")
    api.POST("/generate/token", handlers.CreateAPIToken)
    v1 := api.Group("/v1")
    v1.Use(middleware.AuthenticationMiddleware())
	v1.POST("/user/update", handlers.UpdateUserData)
	v1.POST("/product/transactions/store", handlers.StoreProductTransactions)
    v1.GET("/product/recommendations", handlers.GetRecommendations)
    v1.GET("/api/v1/product/get/all", handlers.GetProducts)
    r.Run(":8080")
}
