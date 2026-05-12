FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o payment-service ./cmd/service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/payment-service .
EXPOSE 50052
CMD ["./payment-service"]
