package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/types"
	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func AddProduct(c *gin.Context) {
	var product types.Product
	err := json.NewDecoder(c.Request.Body).Decode(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	query := `MATCH(i:Index {name: "product_index"})
        SET i.value = i.value + 1
        CREATE(p:Product {id: i.value, name: $name, description: $description, price: $price}),
        (p)-[:HAS_ALLERGY]->(a:Allergens {type: $allergens}),
        (p)-[:GENDER]->(g:Gender {type: $gender}) set p.textEmbedding = $embeddings
        return p.id as id`
	params := map[string]interface{}{
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"allergens":   product.Allergens,
		"gender":      product.Gender,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var products []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		products = append(products, p.AsMap())
	}
	c.JSON(http.StatusCreated, gin.H{"results": products})
}

func EditProduct(c *gin.Context) {
	var product types.Product
	err := json.NewDecoder(c.Request.Body).Decode(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	query := `MATCH (p:Product {id: $id}), (p)-->(a:Allergens), (p)-->(g:Gender)
        SET p.name = $name, p.description = $description, p.price = $price,
        p.textEmbedding = $embeddings,
        a.type = $allergens, g.type = $gender
        return distinct p.id as id`
	params := map[string]interface{}{
		"id":          product.ID,
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"allergens":   product.Allergens,
		"gender":      product.Gender,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var products []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		products = append(products, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"results": products})
}

func GetRecommendations(c *gin.Context) {
	var recquery types.RecommendationQuery
	err := json.NewDecoder(c.Request.Body).Decode(&recquery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isuserexist := utils.CheckIfUserExistsInNeo4J(recquery.UserIc)
	if !isuserexist {
		data := utils.GetUserDataFromPgV2(recquery.UserIc)
		result := utils.StoreUserData(data)
		fmt.Println(result)
	}
	queryVector := utils.GetEmbeddings(recquery.Query)
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
    CALL db.index.vector.queryNodes('product_text_embeddings', $limit, $queryVector)
    YIELD node AS product, score
    WHERE score > 0.65
    MATCH (product)-[:HAS_ALLERGY]->(a:Allergens),
          (product)-[:GENDER]->(g:Gender),
          (product)-[:IS_AFFILIATED_WITH]->(af:Affiliations),
          (u:User {id: $userId})-[:HAS_ALLERGY]->(userAllergen:Allergens),
          (u)-[:GENDER]->(userGender:Gender)
    WHERE (a.type = "Not-Known" OR a.type <> userAllergen)
          AND (g.type = userGender  OR g.type = "Unisex")
          AND af.id = $affiliationID
    RETURN product.name AS name, product.description AS description, product.price AS price, score
    `
	params := map[string]interface{}{
		"limit":         recquery.Limit,
		"queryVector":   queryVector,
		"userId":        recquery.UserIc,
		"affiliationID": recquery.AffiliationID,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var recommendations []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		recommendations = append(recommendations, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

func GetProducts(c *gin.Context) {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	log.Println(driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`
    MATCH (p:Product), (p)-->(a:Allergens), (p)-->(g:Gender)
    RETURN distinct p.id as id, p.name as name,p.description as description,p.price as
    price, a.type as allergens, g.type as gender order by p.id DESC
    `
	params := map[string]interface{}{}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var products []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		products = append(products, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

func StoreProductTransactions(c *gin.Context) {
	var order types.Order
	err := json.NewDecoder(c.Request.Body).Decode(&order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	log.Println(driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`
    UNWIND $product_transactions AS pt
      MATCH(u:User {id: $user_id})
      MATCH(p:Product {id: pt.product_id})
      MERGE (u)-[t:TRANSACTED]->(p)
      set t.order_id = $order_id, t.quantity = pt.quantity
      RETURN p.id, t.order_id, t.quantity,  u.id
    `
	params := map[string]interface{}{
		"order_id":             order.ID,
		"user_id":              order.UserID,
		"product_transactions": order.ProductTransactions,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var productTransactions []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		productTransactions = append(productTransactions, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"results": productTransactions})
}

func GetAffiliations(c *gin.Context) {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	log.Println(driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`
    MATCH(af:Affiliations) return distinct af.id as id, af.name as name;
    `
	params := map[string]any{}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var affiliations []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		affiliations = append(affiliations, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"affiliations": affiliations})
}

func StoreWooCommerceProducts(c *gin.Context) {
	ctx := context.Background()

	// 1. Fetch products from WooCommerce API
	baseUrl := os.Getenv("WOOCOMMERCE_PRODUCT_API")
	consumerKey := os.Getenv("WOOCOMMERCE_CONSUMER_KEY")
	consumerSecret := os.Getenv("WOOCOMMERCE_CONSUMER_SECRET")
	apiUrl, err := url.Parse(baseUrl)
	if err != nil {
		fmt.Printf("Error parsing base URL: %v\n", err)
		return
	}
	query := apiUrl.Query()
	query.Set("consumer_key", consumerKey)
	query.Set("consumer_secret", consumerSecret)
	query.Set("per_page", "100")
	apiUrl.RawQuery = query.Encode()

	finalApiUrl := apiUrl.String()
	response, err := http.Get(finalApiUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products from WooCommerce API: " + err.Error()}) // Include error details
		return
	}
	defer response.Body.Close()

	// 2. Decode the JSON response
	var payload types.WooCommerceProductQuery
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode WooCommerce products: " + err.Error()}) // Include error details
		return
	}

	// 3. Connect to Neo4j (Moved outside the loop)
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Neo4j: " + err.Error()}) // Include error details
		return
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	SecretID := payload.SecretID
	Secret := payload.Secret

	// 4. Store products in Neo4j
	for _, product := range payload.Products {
		// ... (Get embeddings)
		textToEmbed := product.Description + " " + product.ShortDescription
		// Get embeddings
		productEmbeddings := utils.GetEmbeddings(textToEmbed)
		// Check if embedding retrieval was successful
		if len(productEmbeddings) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get embeddings for product"})
			return
		}
		// Construct query with individual product parameters
		query := `
			MATCH(s:Site {secretID: $secretID, secret: $secret})
            CREATE(p:Product {
                id: $id,
                name: $name,
                description: $description,
                short_description: $short_description,
                price: $price,
                permalink: $permalink,
                featured_image: $featured_image,
                textEmbedding: $embeddings
            })
			CREATE (p)-[:BELONGS_TO]->(s)
            RETURN p.id AS id
        `
		params := map[string]interface{}{
			"secretID":          SecretID,
			"secret":            Secret,
			"id":                product.ID,
			"name":              product.Name,
			"description":       product.Description,
			"short_description": product.ShortDescription,
			"price":             product.Price,
			"permalink":         product.Permalink,
			"featured_image":    product.FeaturedImage,
			"embeddings":        productEmbeddings,
		}

		// Execute query for each product
		result, err := session.Run(ctx, query, params)
		if err != nil {
			log.Printf("Error storing product %d: %s", product.ID, err.Error())
			continue // Skip to next product if error occurs
		}

		// Optionally: Log created product ID
		if result.Next(ctx) {
			createdProductID, _ := result.Record().Get("id")
			log.Printf("Created product with ID: %v", createdProductID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Products stored successfully"})
}

func HandleAddProductWebhook(c *gin.Context) {
	ctx := context.Background()

	var payload types.WooCommerceProductQuery

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	// Connect to Neo4j
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Printf("Error connecting to Neo4j: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	SecretID := payload.SecretID
	Secret := payload.Secret

	// Construct the Cypher query
	for _, product := range payload.Products {
		// ... (Get embeddings)
		textToEmbed := product.Description + " " + product.ShortDescription
		// Get embeddings
		productEmbeddings := utils.GetEmbeddings(textToEmbed)
		// Check if embedding retrieval was successful
		if len(productEmbeddings) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get embeddings for product"})
			return
		}
		// Construct query with individual product parameters
		query := `
			MATCH(s:Site {secretID: $secretID, secret: $secret})
            CREATE(p:Product {
	            id: $id,
	            name: $name,
	            description: $description,
	            short_description: $short_description,
	            price: $price,
	            permalink: $permalink,
	            featured_image: $featured_image,
	            textEmbedding: $embeddings
            })
			CREATE (p)-[:BELONGS_TO]->(s)
            RETURN p.id AS id
        `
		params := map[string]interface{}{
			"secretID":          SecretID,
			"secret":            Secret,
			"id":                product.ID,
			"name":              product.Name,
			"description":       product.Description,
			"short_description": product.ShortDescription,
			"price":             product.Price,
			"permalink":         product.Permalink,
			"featured_image":    product.FeaturedImage,
			"embeddings":        productEmbeddings,
		}

		// Execute query for each product
		result, err := session.Run(ctx, query, params)
		if err != nil {
			log.Printf("Error storing product %d: %s", product.ID, err.Error())
			continue // Skip to next product if error occurs
		}

		// Optionally: Log created product ID
		if result.Next(ctx) {
			createdProductID, _ := result.Record().Get("id")
			log.Printf("Created product with ID: %v", createdProductID)
		}
	}

}

func HandleProductUpdateWebhook(c *gin.Context) {
	ctx := context.Background()

	var payload types.WooCommerceProductQuery

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	// Get embeddings for the product
	textToEmbed := payload.Description + " " + payload.ShortDescription
	productEmbeddings := utils.GetEmbeddings(textToEmbed)

	// Connect to Neo4j
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Printf("Error connecting to Neo4j: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)

	SecretID := payload.SecretID
	Secret := payload.Secret

	// Construct the Cypher query
	for _, product := range payload.Products {
		// ... (Get embeddings)
		textToEmbed := product.Description + " " + product.ShortDescription
		// Get embeddings
		productEmbeddings := utils.GetEmbeddings(textToEmbed)
		// Check if embedding retrieval was successful
		if len(productEmbeddings) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get embeddings for product"})
			return
		}
		// Construct query with individual product parameters
		query := `
			MATCH (s:Site {id: $site_id})
			MATCH (p:Product {id: $id})-[r:BELONGS_TO]->(s)
			SET p = {
			   id: $id,
			   name: $name,
			   description: $description,
			   short_description: $short_description,
			   price: $price,
			   permalink: $permalink,
			   featured_image: $featured_image,
			   textEmbedding: $embeddings
			}
			RETURN p.id AS id, s.id AS site_id
        `
		params := map[string]interface{}{
			"secretID":          SecretID,
			"secret":            Secret,
			"id":                product.ID,
			"name":              product.Name,
			"description":       product.Description,
			"short_description": product.ShortDescription,
			"price":             product.Price,
			"permalink":         product.Permalink,
			"featured_image":    product.FeaturedImage,
			"embeddings":        productEmbeddings,
		}

		// Execute query for each product
		result, err := session.Run(ctx, query, params)
		if err != nil {
			log.Printf("Error updating product %d: %s", product.ID, err.Error())
			continue // Skip to next product if error occurs
		}

		// Optionally: Log created product ID
		if result.Next(ctx) {
			createdProductID, _ := result.Record().Get("id")
			log.Printf("Updated product with ID: %v", createdProductID)
		}
	}
}

func HandleProductDeleteWebhook(c *gin.Context) {
	ctx := context.Background()

	var payload types.WooCommerceProductQuery

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	// Connect to Neo4j
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Printf("Error connecting to Neo4j: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)

	for _, product := range payload.Products {
		// Construct the Cypher query to delete the product
		query := `
			MATCH (p:Product {id: $id})
			DETACH DELETE p
		`

		params := map[string]interface{}{
			"id": product.ID,
		}

		// Execute the query
		_, err = session.Run(ctx, query, params)
		if err != nil {
			log.Printf("Error deleting product in Neo4j: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product in database"})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Product with ID %d deleted", payload.ID)})
	}
}

func GetRecommendationsWooCommerce(c *gin.Context) {
	var recquery types.WooCommerceRecommendationQuery
	fmt.Println(c.Request.Body)
	err := json.NewDecoder(c.Request.Body).Decode(&recquery)
	fmt.Println(recquery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// isuserexist := utils.CheckIfUserExistsInNeo4J(recquery.UserId)
	// if !isuserexist{
	//  data := utils.GetUserDataFromPg(recquery.UserId)
	//  result := utils.StoreUserData(data)
	// fmt.Println(result)
	// }
	diagnosis, err := utils.GetUserDiagnosisFromIc(recquery.UserData.IC, recquery.NDiagnosis)
	fmt.Println("Diagnosis", diagnosis)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	combinedDiagnosis := ""
	if len(diagnosis) > 0 {
		combinedDiagnosis += " " + strings.Join(diagnosis, " ")
	}
	fmt.Println(combinedDiagnosis)
	queryVector := utils.GetEmbeddings(combinedDiagnosis)
	fmt.Println(queryVector)
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(
		os.Getenv("NEO4J_URI"),
		neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"),
			os.Getenv("NEO4J_PASSWORD"), "",
		))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`
    CALL db.index.vector.queryNodes('product_text_embeddings', $limit, $queryVector)
    YIELD node AS product, score
    WHERE score > $score_threshold
    RETURN product.id as product_id, product.name as product_name, score
    `
	params := map[string]interface{}{
		"limit":           recquery.Limit,
		"queryVector":     queryVector,
		"score_threshold": recquery.Score,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	fmt.Println(results)
	var recommendations []map[string]any
	for _, p := range results.([]*neo4j.Record) {
		recommendations = append(recommendations, p.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}
