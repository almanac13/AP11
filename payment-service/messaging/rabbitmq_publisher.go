package messaging

import (
	"encoding/json"
	"log"
	"payment-service/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type RabbitMQPublisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

func NewRabbitMQPublisher(url string, queue string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQPublisher{
		conn:  conn,
		ch:    ch,
		queue: queue,
	}, nil
}

func (p *RabbitMQPublisher) PublishPaymentCompleted(payment domain.Payment) error {
	event := PaymentCompletedEvent{
		EventID:       payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		CustomerEmail: "user@example.com",
		Status:        payment.Status,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.ch.Publish(
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)

	if err == nil {
		log.Printf("payment event published: order_id=%s status=%s", event.OrderID, event.Status)
	}

	return err
}

func (p *RabbitMQPublisher) Close() {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
