package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	apiUrl := "https://ecomm.teleme.co/recommendation/wp-json/wc/v3/products?consumer_key=ck_72e1f247b373d3f4e677c357f6f5068a5810d683&consumer_secret=cs_5ce5c9e6c592b9852ec8be66ac3f31d658966372"
	response, err := http.Get(apiUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products from WooCommerce API: " + err.Error()}) // Include error details
		return
	}
	defer response.Body.Close()

	// 2. Decode the JSON response
	var wooProducts []types.WooCommerceProduct
	if err := json.NewDecoder(response.Body).Decode(&wooProducts); err != nil {
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

	// 4. Store products in Neo4j
	for _, product := range wooProducts {
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
            CREATE(p:Product {
                id: $id, 
                name: $name, 
                description: $description,
                short_description: $short_description,
                textEmbedding: $embeddings
            })
            RETURN p.id AS id
        `
		params := map[string]interface{}{
			"id":                product.ID,
			"name":              product.Name,
			"description":       product.Description,
			"short_description": product.ShortDescription,
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

	var payload types.WooCommerceProduct

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

	// Construct the Cypher query
	query := `
    CREATE (p:Product {
        id: $id,
        name: $name,
        description: $description,
        short_description: $short_description,
        textEmbedding: $embeddings
    })
    RETURN p.id AS id
    `
	params := map[string]interface{}{
		"id":                payload.ID,
		"name":              payload.Name,
		"description":       payload.Description,
		"short_description": payload.ShortDescription,
		"embeddings":        productEmbeddings,
	}

	// Execute the query
	result, err := session.Run(ctx, query, params)
	if err != nil {
		log.Printf("Error creating product in Neo4j: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product in database"})
		return
	}

	// Optionally, log the created product ID or return it in the response
	if result.Next(ctx) {
		createdProductID, _ := result.Record().Get("id")
		log.Printf("Created product with ID: %v", createdProductID)
		c.JSON(http.StatusCreated, gin.H{"message": "Product created", "product_id": createdProductID})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product creation failed"})
	}
}

func HandleProductUpdateWebhook(c *gin.Context) {
	ctx := context.Background()

	var payload types.WooCommerceProduct

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

	// Construct the Cypher query to update the product
	query := `
        MATCH (p:Product {id: $id})
        SET p.name = $name,
            p.description = $description,
            p.short_description = $short_description,
            p.textEmbedding = $embeddings
        RETURN p
    `

	params := map[string]interface{}{
		"id":                payload.ID,
		"name":              payload.Name,
		"description":       payload.Description,
		"short_description": payload.ShortDescription,
		"embeddings":        productEmbeddings,
	}

	// Execute the query
	result, err := session.Run(ctx, query, params)
	if err != nil {
		log.Printf("Error updating product in Neo4j: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product in database"})
		return
	}

	// Optionally, log the updated product or return it in the response
	if result.Next(ctx) {
		updatedProduct, _ := result.Record().Get("p")
		log.Printf("Updated product: %v", updatedProduct)
		c.JSON(http.StatusOK, gin.H{"message": "Product updated", "product": updatedProduct})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
	}
}

func HandleProductDeleteWebhook(c *gin.Context) {
	ctx := context.Background()

	var payload types.WooCommerceProduct

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

	// Construct the Cypher query to delete the product
	query := `
        MATCH (p:Product {id: $id})
        DETACH DELETE p
    `

	params := map[string]interface{}{
		"id": payload.ID,
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

func GetRecommendationsWooCommerce(c *gin.Context) {
	var recquery types.WooCommerceRecommendationQuery
	err := json.NewDecoder(c.Request.Body).Decode(&recquery)
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
    RETURN product.id as product_id, score
    `
	params := map[string]interface{}{
		"limit":       recquery.Limit,
		"queryVector": queryVector,
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
