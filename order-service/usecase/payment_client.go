package usecase

import (
	"context"
	"errors"
	"time"

	paymentv1 "github.com/almanac13/ADP2_asik2_generated/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrPaymentServiceUnavailable = errors.New("payment service unavailable")

type PaymentClient struct {
	client paymentv1.PaymentServiceClient
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &PaymentClient{
		client: paymentv1.NewPaymentServiceClient(conn),
	}, nil
}

func (p *PaymentClient) Pay(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := p.client.ProcessPayment(ctx, &paymentv1.PaymentRequest{
		OrderId:        orderID,
		Amount:         amount,
		IdempotencyKey: orderID,
	})
	if err != nil {
		return "", ErrPaymentServiceUnavailable
	}

	return resp.Status, nil
}
