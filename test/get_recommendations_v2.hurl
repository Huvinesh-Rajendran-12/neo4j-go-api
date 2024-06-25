POST http://0.0.0.0:8080/api/v2/product/recommendations
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnlUaW1lIjoxNzE3NDcwNTc0LCJzZWNyZXRfaWQiOiJLbHd5N2xicmFFU1FiNGdxS085MDVKM0VadVpzQVBjSiIsInNlY3JldF9rZXkiOiJsVGlUMkVWWkRQZzY5U1hIcGlVMjlraVkzckdmdlRLMCJ9.s1fVeQpR1oI2ivdJIqpnfVolNdI3FZw2U4I547Wj64Q
{
    "query": "",
    "limit": 10,
    "user_data": {
        "id": 1,
        "ic_passport": "123456789",
        "email": "default@gmail.com"
      }
}
HTTP 200
[Captures]
results: jsonpath "$"
