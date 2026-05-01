package main

import (
	"log"
	"net"
	"os"

	"payment-service/messaging"
	"payment-service/repository"
	servicegrpc "payment-service/transport/grpc"
	"payment-service/usecase"

	paymentv1 "github.com/almanac13/ADP2_asik2_generated/payment/v1"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("PAYMENT_DB_URL")
	grpcPort := os.Getenv("PAYMENT_GRPC_PORT")

	if dbURL == "" {
		log.Fatal("PAYMENT_DB_URL is required")
	}
	if grpcPort == "" {
		grpcPort = "50052"
	}

	db := repository.NewDB(dbURL)
	repo := repository.NewPaymentRepo(db)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	publisher, err := messaging.NewRabbitMQPublisher(rabbitURL, "payment.completed")
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	uc := usecase.NewPaymentUsecase(repo, publisher)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(servicegrpc.LoggingInterceptor),
	)

	paymentv1.RegisterPaymentServiceServer(server, servicegrpc.NewPaymentServer(uc))

	log.Printf("payment-service gRPC running on :%s", grpcPort)

	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
