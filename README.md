# mytheresa

Mytheresa API
================

Introduction
This is a simple API built to provide product information with discounts applied. The API is built using Go and uses a SQLite database to store product data.

Getting Started
Prerequisites
Docker installed on your system
Docker Compose installed on your system
Building and Running the Application
To build and run the application, navigate to the project directory and run the following command:

**docker compose up**

This will start the API server and make it available at http://localhost:8000.

Running Unit Tests
To run the unit tests, navigate to the project directory and run the following command:

**docker compose run app go test -v ./test**

This will execute the unit tests and display the results.

API Endpoints
GET /products
Returns a list of products with discounts applied.

Query Parameters:
+ category: Filter products by category
+ priceLessThan: Filter products by price (less than or equal to the specified value)