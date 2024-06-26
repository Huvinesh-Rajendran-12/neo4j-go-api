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
	r.SetTrustedProxies([]string{"47.254.238.67", "127.0.0.1", "202.184.216.86"})
	api := r.Group("/api")
	api.POST("/authenticate", handlers.AdminAuthentication)
	api.POST("/generate/token", handlers.CreateAPIToken)
	api.POST("/generate/secrets", handlers.GenerateSecret)
	api.GET("/secrets/get/all", handlers.GetSecrets)
	api.GET("/check/token/expiration", handlers.CheckAPITokenExpirations)
	api.GET("/affiliation/get/all", handlers.GetAffiliations)
	api.POST("/product/store/woocommerce", handlers.StoreWooCommerceProducts)
	api.POST("/product/add/woocommerce/webhook", handlers.HandleAddProductWebhook)
	api.POST("/product/update/woocommerce/webhook", handlers.HandleProductUpdateWebhook)
	api.POST("/product/delete/woocommerce/webhook", handlers.HandleProductDeleteWebhook)
	v1 := api.Group("/v1")
	v1.Use(middleware.AuthenticationMiddleware())
	v1.POST("/user/update", handlers.UpdateUserData)
	v1.POST("/product/transactions/store", handlers.StoreProductTransactions)
	v1.GET("/product/recommendations", handlers.GetRecommendations)
	v1.GET("/product/get/all", handlers.GetProducts)
	v2 := api.Group("/v2")
	v2.Use(middleware.AuthenticationMiddleware())
	v2.POST("/product/recommendations", handlers.GetRecommendationsWooCommerce)
	r.Run(":8080")
}
