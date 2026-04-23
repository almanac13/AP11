package main

import (
	"context"
	"log"
	"time"

	paymentv1 "github.com/almanac13/ADP2_asik2_generated/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := paymentv1.NewPaymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListPayments(ctx, &paymentv1.ListPaymentsRequest{
		Status: "Authorized",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range resp.Payments {
		log.Printf("payment_id=%s order_id=%s status=%s amount=%d", p.PaymentId, p.OrderId, p.Status, p.Amount)
	}
}
