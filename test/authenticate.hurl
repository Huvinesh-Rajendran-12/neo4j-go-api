POST http://127.0.0.1:8080/api/authenticate
Content-Type: application/json
{
    "username": "telemeAdmin",
    "password": "teleme@123"
}
HTTP 200
[Captures]
results: jsonpath "$"
