package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/types"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
    "github.com/Huvinesh-Rajendran-12/neo4j-go-api/utils"
)

func AddProduct(c *gin.Context) {
    var product types.Product
	err := json.NewDecoder(c.Request.Body).Decode(&product)
    fmt.Println(product)
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
    fmt.Println(driver)
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
        "name": product.Name,
        "description": product.Description,
        "price": product.Price,
        "allergens": product.Allergens,
        "gender": product.Gender,
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
    fmt.Println(product)
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
    fmt.Println(driver)
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
    query := `MATCH (p:Product {id: $id}), (p)-->(a:Allergens), (p)-->(g:Gender)
        SET p.name = $name, p.description = $description, p.price = $price,
        p.textEmbedding = $embeddings,
        a.type = $allergens, g.type = $gender
        return distinct p.id as id`
    params := map[string]interface{}{
        "id": product.ID,
        "name": product.Name,
        "description": product.Description,
        "price": product.Price,
        "allergens": product.Allergens,
        "gender": product.Gender,
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

func GetRecommendations(c *gin.Context){
    var recquery types.RecommendationQuery
	err := json.NewDecoder(c.Request.Body).Decode(&recquery)
    fmt.Println(recquery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
    isuserexist := utils.CheckIfUserExistsInNeo4J(recquery.UserId)
    fmt.Println(isuserexist)
    if !isuserexist{
        data := utils.GetUserDataFromPg(recquery.UserId)
        result := utils.StoreUserData(data)
        fmt.Println(result)
    }
    queryVector := utils.GetEmbeddings(recquery.Query)
    fmt.Print(queryVector)
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
    fmt.Println(driver)
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
          (u:User {id: $userId})-[:HAS_ALLERGY]->(userAllergen:Allergens),
          (u)-[:GENDER]->(userGender:Gender)
    WHERE (a.type = "Not-Known" OR a.type <> userAllergen)
          AND (g.type = userGender  OR g.type = "Unisex")
    RETURN product.name AS name, product.description AS description, product.price AS price, score
    `
    params := map[string]interface{}{
        "limit": recquery.Limit,
        "queryVector": queryVector,
        "userId": recquery.UserId,

    }
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
    var recommendations  []map[string]any
    for _, p := range results.([]*neo4j.Record) {
        recommendations = append(recommendations, p.AsMap())
    }
	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

func GetProducts(c *gin.Context){
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
    fmt.Println(driver)
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
    fmt.Println(c.Request.Body)
	err := json.NewDecoder(c.Request.Body).Decode(&order)
    fmt.Println(order)
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
    fmt.Println(driver)
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
    fmt.Println(session)
    query := `
      UNWIND $product_transactions AS pt
      MATCH(u:User {id: $user_id})
      MATCH(p:Product {id: pt.product_id})
      MERGE (u)-[t:TRANSACTED]->(p)
      set t.order_id = $order_id, t.quantity = pt.quantity 
      RETURN p.id, t.order_id, t.quantity,  u.id`
	params := map[string]interface{}{
        "order_id": order.ID,
        "user_id": order.UserID,
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
