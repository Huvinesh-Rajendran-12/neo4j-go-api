GET http://0.0.0.0:8080/api/v1/product/recommendations
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTU3MzI4NDMsInVzZXJfaWQiOjB9.UTTjRxjchLfZ6SM1pkDBGwD3IVZRa6o1tKik6qSC0iU
{
    "user_id": 420,
    "query": "sore throat",
    "limit": 5
}
HTTP 200
[Captures]
results: jsonpath "$"
