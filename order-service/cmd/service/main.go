package main

import (
	"log"
	"net"
	"os"

	"order-service/repository"
	servicegrpc "order-service/transport/grpc"
	httptransport "order-service/transport/http"
	"order-service/usecase"

	ordertrackingv1 "github.com/almanac13/ADP2_asik2_generated/ordertracking/v1"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("ORDER_DB_URL")
	httpPort := os.Getenv("ORDER_HTTP_PORT")
	grpcPort := os.Getenv("ORDER_GRPC_PORT")
	paymentAddr := os.Getenv("PAYMENT_GRPC_ADDR")

	if dbURL == "" {
		log.Fatal("ORDER_DB_URL is required")
	}
	if httpPort == "" {
		httpPort = "8080"
	}
	if grpcPort == "" {
		grpcPort = "50051"
	}
	if paymentAddr == "" {
		log.Fatal("PAYMENT_GRPC_ADDR is required")
	}

	db := repository.NewDB(dbURL)
	repo := repository.NewOrderRepo(db)

	paymentClient, err := usecase.NewPaymentClient(paymentAddr)
	if err != nil {
		log.Fatal(err)
	}

	uc := usecase.NewOrderUsecase(repo, paymentClient)
	handler := httptransport.NewOrderHandler(uc)

	go func() {
		r := gin.Default()
		r.POST("/orders", handler.CreateOrder)
		r.GET("/orders/:id", handler.GetOrder)
		r.PATCH("/orders/:id/cancel", handler.CancelOrder)
		r.GET("/orders2/customer/:customer_id", handler.GetOrdersByCustomer)

		log.Printf("order-service HTTP running on :%s", httpPort)

		if err := r.Run(":" + httpPort); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	ordertrackingv1.RegisterOrderTrackingServiceServer(
		grpcServer,
		servicegrpc.NewOrderTrackingServer(repo),
	)

	log.Printf("order-service gRPC running on :%s", grpcPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
