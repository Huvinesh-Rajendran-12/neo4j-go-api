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
)

func UpdateUserData(c *gin.Context) {
	var user types.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
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
	query := `MATCH(u:User {id: $id}) SET u.name = $name, u.age = $age, u.dob = date({year: $year, month: $month, day: $day}) RETURN u`
	params := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"age":   user.Age,
		"year":  user.DOB.Year,
		"month": user.DOB.Month,
		"day":   user.DOB.Day,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
    var persons []map[string]any
	for _, person := range results.([]*neo4j.Record) {
        persons = append(persons, person.AsMap())
	}
	c.JSON(http.StatusOK, gin.H{"message": persons})
}
