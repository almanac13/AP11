# Assignment 4 - Performance Optimization & External Integrations

## Implemented features

### 1. Redis cache-aside in Order Service
`GET /orders/:id` first checks Redis using key `order:{id}`. If Redis has the order, the service returns the cached value. If Redis does not have the order, the service reads from PostgreSQL and stores the result in Redis with TTL from `CACHE_TTL`.

### 2. Cache invalidation
When an order status changes after payment or cancellation, the Order Service deletes `order:{id}` from Redis. This prevents stale statuses such as `Pending` after the database has already changed to `Paid` or `Cancelled`.

### 3. Background notification worker
The Notification Service consumes `payment.completed` messages from RabbitMQ. Email sending is outside the request path, so the API does not wait for slow providers.

### 4. Provider adapter pattern
Notification Service uses the `EmailSender` interface. `PROVIDER_MODE=SIMULATED` starts a mock provider with latency and random failures. `PROVIDER_MODE=REAL` starts an SMTP placeholder adapter.

### 5. Retry + exponential backoff
If sending fails, the worker retries with exponential backoff: 2s, 4s, 8s. Retry count is configured through `MAX_RETRIES`.

### 6. Redis idempotency
Before sending a notification, the worker checks Redis key `notification:payment:{payment_id}`. If the key already exists, the duplicate job is skipped. After successful send, the key is saved for 24 hours.

### 7. Bonus Redis rate limiter
Order Service has Redis middleware limiting requests by client IP. If request count is higher than `RATE_LIMIT` per minute, it returns HTTP 429.

## Run

```bash
docker compose up --build
```

## Test order creation

```bash
curl -X POST http://localhost:8080/orders ^
  -H "Content-Type: application/json" ^
  -d "{\"customer_id\":\"c1\",\"item_name\":\"Book\",\"amount\":2000}"
```

## Test cache

Run twice with the returned order id:

```bash
curl http://localhost:8080/orders/ORDER_ID
curl http://localhost:8080/orders/ORDER_ID
```

Expected logs:

```text
[Redis] cache MISS for order_id=...
[Redis] cached order_id=... ttl=5m
[Redis] cache HIT for order_id=...
```

## Test notification worker

Create orders several times. Because the provider is simulated, sometimes logs show failures and retries:

```text
[Worker] attempt 1 failed: simulated provider temporary failure; retrying in 2s
[Worker] attempt 2 failed: simulated provider temporary failure; retrying in 4s
[Worker] notification processed payment_id=...
```

## Defense explanation

This project improves the original microservices by adding Redis caching to reduce database pressure, RabbitMQ background jobs to avoid blocking API requests, an adapter pattern for external email providers, retry with exponential backoff for temporary failures, and Redis idempotency to prevent duplicate notifications.
