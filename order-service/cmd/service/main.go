package main

import (
	"order-service/repository"
	"order-service/transport/http"
	"order-service/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	db := repository.NewDB("postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable")

	repo := repository.NewOrderRepo(db)
	paymentClient := usecase.NewPaymentClient("http://localhost:8081")
	usecase := usecase.NewOrderUsecase(repo, paymentClient)
	handler := http.NewOrderHandler(usecase)

	r := gin.Default()

	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)
	r.GET("/orders2/customer/:customer_id", handler.GetOrdersByCustomer)

	r.Run(":8080")
}
