# Assignment 2 — gRPC Migration & Contract-First Development

## Overview
This project is an upgraded version of Assignment 1.  
The system keeps the external REST API for clients, but internal communication between services has been migrated from REST to gRPC.

Services:
- **Order Service**
- **Payment Service**

Main improvements:
- internal service-to-service communication via **gRPC**
- **Protocol Buffers** as contracts
- **Contract-First** workflow with separate repositories
- **Server-side streaming** for order tracking
- configuration through **environment variables**
- **Unary interceptor** for logging on Payment Service
- generated protobuf code managed through **GitHub Actions**

---

## Repositories

### Main project repository
- `AP11`

### Proto repository
- `ADP2_asik2_protos`

### Generated code repository
- `ADP2_asik2_generated`

---

## Architecture

### External communication
Client → REST → Order Service

### Internal communication
Order Service → gRPC → Payment Service

### Streaming
Stream Client → gRPC stream → Order Service

---

## Features

### 1. REST API for external users
Order Service still exposes REST endpoints:
- `POST /orders`
- `GET /orders/:id`
- `PATCH /orders/:id/cancel`

### 2. gRPC communication between services
Order Service acts as a **gRPC client** and calls Payment Service using:
- `ProcessPayment(PaymentRequest) returns (PaymentResponse)`

Payment Service acts as a **gRPC server**.

### 3. Server-side streaming
Order Service also exposes:
- `SubscribeToOrderUpdates(OrderRequest) returns (stream OrderStatusUpdate)`

A separate stream client can subscribe to order status updates and receive them in real time.

### 4. Contract-First workflow
Protocol Buffers are stored in a separate repository.  
Generated Go code is stored in another separate repository.  
Generation is automated through GitHub Actions.

### 5. Interceptor
Payment Service includes a unary interceptor that logs:
- gRPC method name
- request duration
- error value

---

## Technologies
- Go
- Gin
- gRPC
- Protocol Buffers
- PostgreSQL
- GitHub Actions

---

## Environment Variables

### Order Service
Create `order-service/.env`

```env
ORDER_DB_URL=postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable
ORDER_HTTP_PORT=8080
ORDER_GRPC_PORT=50051
PAYMENT_GRPC_ADDR=localhost:50052