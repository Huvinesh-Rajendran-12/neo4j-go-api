package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"crypto/rand"
	"encoding/hex"
	"github.com/Huvinesh-Rajendran-12/neo4j-go-api/types"
	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GenerateToken generates a JWT token with the user ID as part of the claims
func GenerateToken(secretID string, secretKey string) (map[string]interface{}, error) {
	secretkeystr := os.Getenv("SECRET_KEY")
	var secretkeybytes = []byte(secretkeystr)
	claims := jwt.MapClaims{}
	claims["secret_id"] = secretID
	claims["secret_key"] = secretKey
	expTime := time.Now().Add(time.Hour * 1).Unix()
	claims["expiryTime"] = expTime // Token valid for 1 hour

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(secretkeybytes)
	data := map[string]interface{}{
		"token":   signedString,
		"expTime": expTime,
	}
	return data, err
}

// VerifyToken verifies a token JWT validate
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	secretkeystr := os.Getenv("SECRET_KEY")
	var secretKey = []byte(secretkeystr)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("Invalid signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	return claims, nil
}

func IsTokenExpired(tokenString string) bool {
	secretkeystr := os.Getenv("SECRET_KEY")
	secretKey := []byte(secretkeystr)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("Invalid signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false
	}
	expiryTime := claims["expiryTime"].(float64)
	expiryTimeUnix := int64(expiryTime)
	expiryTimeT := time.Unix(expiryTimeUnix, 0)
	return expiryTimeT.After(time.Now())
}

// parseDate parses the date of time module into year, month, day
func ParseDate(date time.Time) (int, int, int) {
	year := date.Year()
	month := int(date.Month())
	day := date.Day()
	return year, month, day
}

// getUserDataFromPg returns the user data from the database
func GetUserDataFromPg(id int) map[string]interface{} {
	host := os.Getenv("POSTGRES_HOST")
	password := os.Getenv("POSTGRES_PASSWORD")
	username := os.Getenv("POSTGRES_USER")
	port := os.Getenv("POSTGRES_PORT")
	database := os.Getenv("POSTGRES_DB")
	db_url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	conn, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	var user_id int
	var email string
	var name string
	var gender string
	var date_of_birth time.Time
	var latitude float64
	var longitude float64
	var allergy string
	query := `SELECT 
    id, 
    COALESCE(email, 'default_email@example.com') AS email, 
    COALESCE(name, 'Unknown') AS name, 
    COALESCE(gender, 'Unknown') AS gender, 
    COALESCE(date_of_birth, '1000-01-01') AS date_of_birth, 
    COALESCE(latitude, 0.0) AS latitude, 
    COALESCE(longitude, 0.0) AS longitude 
    FROM users 
    WHERE email = $1;`
	err = conn.QueryRow(context.Background(), query, email).
		Scan(&user_id, &email, &name, &gender, &date_of_birth, &latitude, &longitude)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	allergy_query := "select COALESCE(name, 'Unknown') from allergies where user_id=$1"
	allergy_err := conn.QueryRow(context.Background(), allergy_query, id).Scan(&allergy)
	if allergy_err != nil {
		if allergy_err == sql.ErrNoRows {
			allergy = ""
		} else {
			allergy = ""
		}
	}
	age := int(time.Since(date_of_birth).Hours() / 24 / 365)
	year, month, day := ParseDate(date_of_birth)
	data := map[string]interface{}{
		"id":        user_id,
		"email":     email,
		"name":      name,
		"age":       age,
		"gender":    gender,
		"latitude":  latitude,
		"longitude": longitude,
		"allergy":   allergy,
		"year":      year,
		"month":     month,
		"day":       day,
	}
	return data
}

func GetUserDataFromPgV2(email string) map[string]interface{} {
	host := os.Getenv("POSTGRES_HOST")
	password := os.Getenv("POSTGRES_PASSWORD")
	username := os.Getenv("POSTGRES_USER")
	port := os.Getenv("POSTGRES_PORT")
	database := os.Getenv("POSTGRES_DB")
	db_url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	conn, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	var user_id int
	var name string
	var gender string
	var date_of_birth time.Time
	var latitude float64
	var longitude float64
	var allergy string
	query := `SELECT 
    id, 
    COALESCE(email, 'default_email@example.com') AS email, 
    COALESCE(name, 'Unknown') AS name, 
    COALESCE(gender, 'Unknown') AS gender, 
    COALESCE(date_of_birth, '1000-01-01') AS date_of_birth, 
    COALESCE(latitude, 0.0) AS latitude, 
    COALESCE(longitude, 0.0) AS longitude 
    FROM users 
    WHERE email = $1;`
	err = conn.QueryRow(context.Background(), query, email).
		Scan(&user_id, &email, &name, &gender, &date_of_birth, &latitude, &longitude)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	age := int(time.Since(date_of_birth).Hours() / 24 / 365)
	year, month, day := ParseDate(date_of_birth)
	data := map[string]interface{}{
		"id":        user_id,
		"email":     email,
		"name":      name,
		"age":       age,
		"gender":    gender,
		"latitude":  latitude,
		"longitude": longitude,
		"allergy":   allergy,
		"year":      year,
		"month":     month,
		"day":       day,
	}
	return data
}

func GetUserDiagnosisFromIc(ic_passport string) ([]string, error) {
	host := os.Getenv("POSTGRES_HOST")
	password := os.Getenv("POSTGRES_PASSWORD")
	username := os.Getenv("POSTGRES_USER")
	port := os.Getenv("POSTGRES_PORT")
	database := os.Getenv("POSTGRES_DB")
	db_url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	conn, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	query := `SELECT c.diagnosis
  FROM consultations c
  JOIN users u ON c.user_id = u.id
  WHERE u.ic = $1
  ORDER BY c.created_at DESC LIMIT 3;`
	rows, err := conn.Query(context.Background(), query, ic_passport)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	var diagnoses []string
	for rows.Next() {
		var diagnosis string
		if err := rows.Scan(&diagnosis); err != nil {
			return nil, fmt.Errorf("row scan failed: %v", err)
		}
		diagnoses = append(diagnoses, diagnosis)
	}

	return diagnoses, nil
}

func GetUserDiagnosisFromEmail(email string) ([]string, error) {
	host := os.Getenv("POSTGRES_HOST")
	password := os.Getenv("POSTGRES_PASSWORD")
	username := os.Getenv("POSTGRES_USER")
	port := os.Getenv("POSTGRES_PORT")
	database := os.Getenv("POSTGRES_DB")
	db_url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	conn, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	query := `SELECT c.diagnosis
  FROM consultations c
  JOIN users u ON c.user_id = u.id
  WHERE u.email = $1
  ORDER BY c.created_at DESC LIMIT 3;`
	rows, err := conn.Query(context.Background(), query, email)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	var diagnoses []string
	for rows.Next() {
		var diagnosis string
		if err := rows.Scan(&diagnosis); err != nil {
			return nil, fmt.Errorf("row scan failed: %v", err)
		}
		diagnoses = append(diagnoses, diagnosis)
	}

	return diagnoses, nil
}

// checkIfUserExistsInNeo4J return true if user exists in graph database
func CheckIfUserExistsInNeo4J(ic_passport string) bool {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query := `MATCH(u:User {ic_passport: $ic_passport}) RETURN u`
	params := map[string]interface{}{
		"ic_passport": ic_passport,
	}
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, params)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var users []map[string]any
	for _, person := range results.([]*neo4j.Record) {
		users = append(users, person.AsMap())
	}
	if len(users) > 0 {
		return true
	} else {
		return false
	}
}

func StoreUserData(userData map[string]interface{}) []map[string]interface{} {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URI"), neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		return []map[string]interface{}{}
	}
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DB")})
	defer session.Close(ctx)
	query :=
		`
    CREATE(u:User {id: $id, name: $name, age: $age, email: $email, latitude: $latitude, longitude: $longitude, 
    dob: date({year: $year, month: $month, day: $day})}), 
    (u)-[:HAS_ALLERGY]->(a: Allergens {type: $allergy}), (u)-[:GENDER]->(g: Gender {type: $gender}) return u, a, g
    `
	results, _ := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, query, userData)
			records, _ := result.Collect(ctx)
			return records, nil
		})
	var persons []map[string]any
	for _, person := range results.([]*neo4j.Record) {
		persons = append(persons, person.AsMap())
	}
	return persons
}

// getEmbeddings returns the embeddings of a text
func GetEmbeddings(text string) []float64 {
	embeddings_api := os.Getenv("EMBEDDINGS_API")
	url := embeddings_api + "?" + "text=" + url.QueryEscape(text)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var embeddingsresp types.EmbeddingResp
	err = json.NewDecoder(resp.Body).Decode(&embeddingsresp)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}
	return embeddingsresp.Embeddings
}

func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
