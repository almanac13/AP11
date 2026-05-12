package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"time"

	"notification-service/provider"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

func main() {
	ctx := context.Background()

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	providerMode := getEnv("PROVIDER_MODE", "SIMULATED")
	maxRetries := getEnvInt("MAX_RETRIES", 3)

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	sender := buildEmailSender(providerMode)

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	queueName := "payment.completed"

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		for msg := range msgs {
			if err := processMessage(ctx, redisClient, sender, msg.Body, maxRetries); err != nil {
				log.Printf("[Worker] job failed permanently: %v", err)
				msg.Nack(false, false)
				continue
			}

			msg.Ack(false)
		}
	}()

	log.Printf("notification-service worker listening on queue=%s provider=%s", queueName, providerMode)

	<-stop
	log.Println("notification-service shutting down")
}

func processMessage(ctx context.Context, redisClient *redis.Client, sender provider.EmailSender, body []byte, maxRetries int) error {
	var event PaymentCompletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}

	idempotencyKey := "notification:payment:" + event.EventID

	processed, err := redisClient.Exists(ctx, idempotencyKey).Result()
	if err != nil {
		return err
	}
	if processed == 1 {
		log.Printf("[Worker] duplicate notification skipped payment_id=%s", event.EventID)
		return nil
	}

	subject := "Payment completed"
	message := fmt.Sprintf("Order %s payment completed. Amount: %d. Status: %s", event.OrderID, event.Amount, event.Status)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := sender.Send(event.CustomerEmail, subject, message)
		if err == nil {
			if err := redisClient.Set(ctx, idempotencyKey, "sent", 24*time.Hour).Err(); err != nil {
				return err
			}
			log.Printf("[Worker] notification processed payment_id=%s order_id=%s", event.EventID, event.OrderID)
			return nil
		}

		if attempt == maxRetries {
			return err
		}

		backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		log.Printf("[Worker] attempt %d failed: %v; retrying in %s", attempt, err, backoff)
		time.Sleep(backoff)
	}

	return nil
}

func buildEmailSender(mode string) provider.EmailSender {
	if mode == "REAL" {
		return provider.NewSMTPSender()
	}

	latency, err := time.ParseDuration(getEnv("PROVIDER_LATENCY", "2s"))
	if err != nil {
		latency = 2 * time.Second
	}

	return provider.NewMockSender(latency)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
