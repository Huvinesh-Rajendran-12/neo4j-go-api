POST http://127.0.0.1:8080/api/v1/user/store
{
    "id": 1,
    "name": "Geetha",
    "age": 30,
    "email": "geetha@gmail.com",
    "dob": {
        "year": 1994,
        "month": 4,
        "day": 6
    }
}

HTTP 201
[Captures]
results: jsonpath "$"
