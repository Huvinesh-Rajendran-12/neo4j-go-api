GET http://0.0.0.0:8080/api/v1/product/recommendations
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnlUaW1lIjoxNzE2MTkxOTEzLCJzZWNyZXRfaWQiOiJLbHd5N2xicmFFU1FiNGdxS085MDVKM0VadVpzQVBjSiIsInNlY3JldF9rZXkiOiJsVGlUMkVWWkRQZzY5U1hIcGlVMjlraVkzckdmdlRLMCJ9.xJ2cPEhrk4dce_UXdSph9RQl5zHNoui-te7YOH5FuxQ 
{
    "user_id": 420,
    "query": "sore throat",
    "affiliation_id": 1,
    "limit": 5
}
HTTP 200
[Captures]
results: jsonpath "$"
