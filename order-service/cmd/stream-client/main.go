package main

import (
	"context"
	"io"
	"log"
	"time"

	ordertrackingv1 "github.com/almanac13/ADP2_asik2_generated/ordertracking/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := ordertrackingv1.NewOrderTrackingServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stream, err := client.SubscribeToOrderUpdates(ctx, &ordertrackingv1.OrderRequest{
		OrderId: "61aa08ec-c689-4490-b107-2b791298b865",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("subscribed to order updates")

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("order_id=%s status=%s updated_at=%s", msg.OrderId, msg.Status, msg.UpdatedAt.AsTime())
	}
}
