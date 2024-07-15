package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Function for logging in
func CreateAPIToken(c *gin.Context) {
	var user struct {
		SecretID string `json:"secret_id"`
		Secret   string `json:"secret"`
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

	c.JSON(http.StatusOK, gin.H{"authentication": data})
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
	log.Println(token)
	data := utils.IsTokenExpired(token.Token)
	c.JSON(http.StatusOK, gin.H{"validity": data})
}

func AdminAuthentication(c *gin.Context) {
	var admin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	// Check request content type (optional)
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content type"})
		return
	}

	// Bind request body to struct
	if err := c.BindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	log.Println(admin)
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query := `MATCH(a:Admin {username: $username}) return a.username as username, a.password as password`
	params := map[string]interface{}{
		"username": admin.Username,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var admins []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		admins = append(admins, p.AsMap())
	}
	admin_ := admins[0]
	log.Println("login admin", admin.Password)
	log.Println("original admin", admin_["password"])
	if !(admin_["password"] == admin.Password) {
		c.JSON(http.StatusForbidden, gin.H{"authenticated": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"authenticated": true})
	}
}

func GenerateSite(c *gin.Context) {
	var data struct {
		Title   string `json:"title"`
		SiteUrl string `json:"site_url"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	secretID, err := utils.GenerateRandomHex(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secret ID"})
		return
	}

	secret, err := utils.GenerateRandomHex(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secret"})
		return
	}

	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`
    MATCH(i:Index {name: "site_index"})
    SET i.value = i.value + 1
    CREATE(s:Site {id: i.value, name: $name, secretID: $secretID,
    secret: $secret, url: $siteUrl }) return s.id as id, s.name as name, s.secretID as secretID, s.secret as secret, s.siteUrl as url
    `
	params := map[string]interface{}{
		"name":     data.Title,
		"siteUrl":  data.SiteUrl,
		"secretID": secretID,
		"secret":   secret,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var secrets []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		secrets = append(secrets, p.AsMap())
	}

	c.JSON(http.StatusOK, gin.H{"results": secrets})
}

func GetSecrets(c *gin.Context) {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`                
    MATCH (s:Secret) RETURN distinct s.id as id, s.name as name,s.secretID as secretID ,s.secret as secret order by s.id DESC
    `
	params := map[string]interface{}{}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var secrets []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		secrets = append(secrets, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"secrets": secrets})
}
