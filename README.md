# Neo4j Go API
This project implements a Go API that interacts with a Neo4j database, demonstrating how to perform CRUD operations and manage relationships in a graph database using Go.

## Features
* CRUD Operations: Create, Read, Update, and Delete nodes in the Neo4j database.
* Relationship Management: Establish and manage relationships between nodes.
* RESTful API: Exposes Neo4j operations through a RESTful API.
* Dockerized: Includes a Dockerfile for easy deployment and containerization.

## Prerequisites
* Go 1.16 or higher
* Neo4j 4.x or higher
* Docker (optional)

## Getting Started
Clone the repository:
```bash
git clone https://github.com/Huvinesh-Rajendran-12/neo4j-go-api.git
```

Set up your Neo4j database and update the connection details in the configuration file.

Install dependencies:
```bash
go mod tidy
```

Run the application:
```bash
go run main.go
```

## API Endpoints

* GET /api/users: Retrieve all users
* GET /api/users/:id: Retrieve a specific user
* POST /api/users: Create a new user
* PUT /api/users/:id: Update a user
* DELETE /api/users/:id: Delete a user

## Docker Support

To run the application using Docker:
Build the Docker image:
```bash
docker build -t neo4j-go-api .
```

Run the container:
```bash
docker run -p 8080:8080 neo4j-go-api
```
