POST http://127.0.0.1:8080/api/v1/product/transactions/store
{
    "id": 1,
    "user_id": 1,
    "product_transactions": [
        {
            "product_id": 1,
            "quantity": 1
        }
    ]
}
HTTP 200
[Captures]
results: jsonpath "$"
