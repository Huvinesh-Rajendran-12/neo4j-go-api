package types

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
	DOB   struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"dob"`
}

type ProductTransaction struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Order struct {
	ID                  int                      `json:"id"`
	UserID              int                      `json:"user_id"`
	ProductTransactions []map[string]interface{} `json:"product_transactions"`
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	Allergens   string `json:"allergens"`
	Gender      string `json:"gender"`
}

type RecommendationQuery struct {
	UserIc        string `json:"user_ic"`
	Query         string `json:"query"`
	AffiliationID int    `json:"affiliation_id"`
	Limit         int    `json:"limit"`
}

type EmbeddingResp struct {
	Embeddings []float64 `json:"embeddings"`
}

type WooCommerceProduct struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Description      string `json:"description"`
	ShortDescription string `json:"short_description"`
	Image            []struct {
		Src string `json:"src"`
	} `json:"images"`
}

type WooCommerceRecommendationQuery struct {
	Limit      int     `json:"limit"`
	Query      string  `json:"query"`
	Score      float64 `json:"score"`
	NDiagnosis int     `json:"n_diagnosis"`
	UserData   struct {
		ID    int    `json:"id"`
		IC    string `json:"ic_passport"`
		Email string `json:"email"`
	} `json:"user_data"`
}
