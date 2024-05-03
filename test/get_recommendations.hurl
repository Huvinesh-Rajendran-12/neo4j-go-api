GET http://0.0.0.0:8080/api/v1/product/recommendations
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTU2MTQ1MDQsInVzZXJfaWQiOjB9.lN-kRyteuAKe4lx4i4kJ7m8CfVd22KRiEpSjfXoMJJA
{
    "user_id": 400,
    "query": "sore throat",
    "limit": 5
}
HTTP 200
[Captures]
results: jsonpath "$"
