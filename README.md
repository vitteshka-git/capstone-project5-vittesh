# Order Management System – Go Microservices

## 1. Project Overview

The Order Management System is a Go-based application that manages customer orders using REST APIs and GraphQL. It follows a microservices architecture with an API Gateway, an Order Management API, and a Product Service.

PostgreSQL is used for persistent order storage, with the database hosted on AWS RDS. JWT authentication protects selected endpoints, and SoapUI is used to test the APIs.

## 2. Technologies Used

- **Programming Language:** Go
- **Web Framework:** Gin
- **Database:** PostgreSQL
- **Database Hosting:** AWS RDS
- **ORM:** GORM
- **API Styles:** REST and GraphQL
- **GraphQL Library:** `github.com/graphql-go/graphql`
- **Authentication:** JWT
- **API Testing:** SoapUI
- **Configuration:** Environment variables using `.env`

## 3. Architecture

The application consists of three main services.

### API Gateway – Port 8080

The API Gateway is the entry point for client requests. It forwards supported requests to the Order Management API.

### Order Management API – Port 8081

This service implements order business logic, REST endpoints, GraphQL queries and mutations, authentication integration, and database access.

### Product Service – Port 8082

The Product Service provides product lookup functionality. The Order Management service uses it to retrieve product information when processing orders.

### Database

PostgreSQL, hosted on AWS RDS, stores order information, including customer ID, product ID, quantity, unit price, total amount, and status.

## 4. Project Structure

The project contains the following main components:

```text
order-management/
├── cmd/
│   ├── gateway/
│   └── product-service/
├── config/
├── controller/
├── database/
├── gateway/
├── graphqlapi/
├── middleware/
├── model/
├── product/
├── repository/
├── service/
├── tests/
├── .env
├── go.mod
├── go.sum
├── main.go
└── README.md
```

### Responsibilities

- `cmd/gateway/` – API Gateway entry point.
- `cmd/product-service/` – Product Service entry point.
- `config/` – Application and database configuration.
- `controller/` – REST request handlers.
- `database/` – Database connection and related setup.
- `gateway/` – Gateway routing and request forwarding logic.
- `graphqlapi/` – GraphQL schema, queries, mutations, and handler.
- `middleware/` – Authentication and HTTP middleware.
- `model/` – Order data structures and request models.
- `product/` – Product Service client.
- `repository/` – Database operations.
- `service/` – Order business logic.
- `tests/` – Test-related files.

## 5. Business Functionalities

The project implements four core order functionalities.

### 5.1 Create Order

**Endpoint:** `POST /api/v1/orders`

Creates an order using customer ID, product ID, and quantity.

Example request:

```json
{
  "customerId": 101,
  "productId": 501,
  "quantity": 2
}
```

The application processes the request and returns the created order, including its generated ID, unit price, total amount, and status.

### 5.2 Get Order

**Endpoint:** `GET /api/v1/orders/{id}`

Retrieves an order by its ID.

Example:

```http
GET /api/v1/orders/4
```

The response includes the order details and timestamps.

### 5.3 List Orders

**Endpoint:** `GET /api/v1/orders`

Retrieves the list of orders.

Example:

```http
GET /api/v1/orders
```

### 5.4 Update Order Status

**Endpoint:** `PUT /api/v1/orders/{id}/status`

Updates the status of an existing order.

Example:

```http
PUT /api/v1/orders/4/status
```

Request body:

```json
{
  "status": "CONFIRMED"
}
```

The application should validate the requested status according to its implementation.

## 6. GraphQL API

The application also exposes a GraphQL endpoint:

```http
POST /graphql
```

Through the API Gateway, the URL is:

```text
http://localhost:8080/graphql
```

### GraphQL Query – List Orders

```json
{
  "query": "query { orders(limit: 20, offset: 0) { id customerId productId quantity unitPrice totalAmount status } }"
}
```

### GraphQL Query – Get Order

```json
{
  "query": "query { order(id: 4) { id customerId productId quantity unitPrice totalAmount status } }"
}
```

### GraphQL Mutation – Create Order

```json
{
  "query": "mutation { createOrder(input: { customerId: 1, productId: 1, quantity: 2 }) { id customerId productId quantity unitPrice totalAmount status } }"
}
```

### GraphQL Mutation – Update Order Status

```json
{
  "query": "mutation { updateOrderStatus(id: 4, status: \"CONFIRMED\") { id status updatedAt } }"
}
```

The GraphQL endpoint is protected by JWT authentication in the current application configuration.

## 7. Authentication

The application provides a login endpoint:

```http
POST /api/v1/auth/login
```

Example URL through the API Gateway:

```text
http://localhost:8080/api/v1/auth/login
```

Submit the credentials expected by the application's login handler. On successful authentication, use the returned JWT token for protected requests.

Add the following HTTP header to protected requests:

```http
Authorization: Bearer YOUR_JWT_TOKEN
```

Do not commit real passwords, JWT secrets, or database credentials to source control.

## 8. Environment Configuration

The application uses a `.env` file for configuration.

Typical configuration keys include:

```dotenv
DB_HOST=your-rds-endpoint
DB_PORT=5432
DB_USER=your-database-user
DB_PASSWORD=your-database-password
DB_NAME=your-database-name
DB_SSLMODE=your-configured-sslmode

PRODUCT_SERVICE_URL=http://localhost:8082
SERVER_PORT=8081
```

Use the exact environment variable names supported by your configuration code. Keep actual credentials in your local `.env` file and exclude that file from version control.

For AWS RDS, ensure that the database instance is available, the hostname is correct, and the security group permits connections from the application's network.

## 9. Prerequisites

Install and configure:

- Go
- PostgreSQL-compatible AWS RDS database
- AWS account and database access, if using the remote RDS instance
- SoapUI
- Git, if required for source control

Ensure that the configured Product Service URL and database connection settings are correct.

## 10. Run the Application

Open separate PowerShell terminals in the project directory.

### Terminal 1 – Start Order Management API

```powershell
cd C:\Users\vittesh.ka\Downloads\124776_Assessment5_Vittesh\order-management
go mod tidy
go run .
```

The API listens on port `8081` by default in the current configuration.

### Terminal 2 – Start API Gateway

```powershell
cd C:\Users\vittesh.ka\Downloads\124776_Assessment5_Vittesh\order-management
go run .\cmd\gateway
```

The gateway listens on port `8080` in the current configuration.

### Terminal 3 – Start Product Service

```powershell
cd C:\Users\vittesh.ka\Downloads\124776_Assessment5_Vittesh\order-management
go run .\cmd\product-service
```

The Product Service listens on port `8082` in the current configuration.

Keep all three terminals running while testing the complete application.

If a port is already occupied, identify the process before stopping it. Avoid running duplicate instances on the same port.

## 11. Health Checks

### Order Management API

Liveness:

```http
GET http://localhost:8081/health/live
```

Readiness:

```http
GET http://localhost:8081/health/ready
```

### Product Service

Liveness:

```http
GET http://localhost:8082/health/live
```

The readiness endpoint checks whether the Order Management API can access its database.

## 12. Testing with SoapUI

Use SoapUI to create HTTP requests and verify the API responses.

For REST requests, select the appropriate HTTP method and URL.

For protected endpoints, add:

```http
Authorization: Bearer YOUR_JWT_TOKEN
```

For JSON requests, add:

```http
Content-Type: application/json
```

For GraphQL, send a JSON body containing the `query` property to the `/graphql` endpoint.

Check the HTTP status code, response body, and request ID when troubleshooting.

## 13. Testing Progress

The following operations have been tested successfully during development:

- **Create Order:** Successful GraphQL mutation; created order ID `4`.
- **Get Order:** Successful REST request for order ID `4`, returning HTTP `200 OK`.
- **List Orders:** GraphQL query returned existing order records successfully.
- **GraphQL Authentication:** The protected GraphQL request succeeded with the configured authentication flow.
- **Database Migration:** GORM migration completed successfully.
- **Service Startup:** The Order Management API, API Gateway, and Product Service reported startup messages.

The REST List Orders and REST Update Order Status operations should be verified separately in SoapUI before marking all four REST business functionalities complete.

## 14. Troubleshooting

### Database connection failure

Check the RDS endpoint, credentials, database name, network connectivity, and inbound security group rules.

### Port already in use

Check the process using the port:

```powershell
netstat -ano | findstr :8081
```

Identify the process before terminating it.

### Unauthorized response

Check that the login succeeded, the JWT is valid, and the request contains the correct `Authorization` header.

### Gateway error

Verify that the gateway is running on port `8080` and that the Order Management API is accessible on port `8081`.

### Product lookup failure

Verify that the Product Service is running on port `8082`, that `PRODUCT_SERVICE_URL` is correct, and that the requested product exists.

## 15. Conclusion

This project demonstrates a Go-based order management system using microservices, REST APIs, GraphQL, JWT authentication, PostgreSQL on AWS RDS, and an API Gateway.

It provides core order operations, separates business logic from database access, and supports API testing with SoapUI.
