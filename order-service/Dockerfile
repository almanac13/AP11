FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o order-service ./cmd/service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-service .
EXPOSE 8080 50051
CMD ["./order-service"]
