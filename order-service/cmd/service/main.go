package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"order-service/cache"
	"order-service/middleware"
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
	redisAddr := os.Getenv("REDIS_ADDR")
	cacheTTLValue := os.Getenv("CACHE_TTL")

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
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if cacheTTLValue == "" {
		cacheTTLValue = "5m"
	}

	cacheTTL, err := time.ParseDuration(cacheTTLValue)
	if err != nil {
		log.Fatalf("invalid CACHE_TTL: %v", err)
	}

	db := repository.NewDB(dbURL)
	repo := repository.NewOrderRepo(db)

	paymentClient, err := usecase.NewPaymentClient(paymentAddr)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := cache.NewRedisClient(redisAddr)
	orderCache := cache.NewRedisOrderCache(redisClient, cacheTTL)

	uc := usecase.NewOrderUsecase(repo, paymentClient, orderCache)
	handler := httptransport.NewOrderHandler(uc)

	go func() {
		r := gin.Default()

		rateLimit, _ := strconv.ParseInt(os.Getenv("RATE_LIMIT"), 10, 64)
		if rateLimit <= 0 {
			rateLimit = 10
		}
		r.Use(middleware.RedisRateLimiter(redisClient, rateLimit, time.Minute))

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
