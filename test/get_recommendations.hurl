GET http://0.0.0.0:8080/api/v1/product/recommendations
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnlUaW1lIjoxNzE0OTg2MzgyLCJzZWNyZXRfaWQiOiJLbHd5N2xicmFFU1FiNGdxS085MDVKM0VadVpzQVBjSiIsInNlY3JldF9rZXkiOiJsVGlUMkVWWkRQZzY5U1hIcGlVMjlraVkzckdmdlRLMCJ9.4ENfOAtz1xTBPp5wn3FRUTGHZxS6KvIY9ayvQDyYqu8
{
    "user_id": 420,
    "query": "sore throat",
    "limit": 5
}
HTTP 200
[Captures]
results: jsonpath "$"
