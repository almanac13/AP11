package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

var processedEvents = struct {
	sync.Mutex
	data map[string]bool
}{
	data: make(map[string]bool),
}

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

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

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		for msg := range msgs {
			var event PaymentCompletedEvent

			if err := json.Unmarshal(msg.Body, &event); err != nil {
				msg.Nack(false, false)
				continue
			}

			processedEvents.Lock()
			if processedEvents.data[event.EventID] {
				processedEvents.Unlock()
				log.Printf("[Notification] Duplicate event skipped: %s", event.EventID)
				msg.Ack(false)
				continue
			}

			log.Printf(
				"[Notification] Sent email to %s for Order #%s. Amount: %d",
				event.CustomerEmail,
				event.OrderID,
				event.Amount,
			)

			processedEvents.data[event.EventID] = true
			processedEvents.Unlock()

			msg.Ack(false)
		}
	}()

	log.Println("notification-service is listening for payment.completed events")

	<-stop
	log.Println("notification-service shutting down")
}
