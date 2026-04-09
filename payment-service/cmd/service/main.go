package main

import (
	"payment-service/repository"
	"payment-service/transport/http"
	"payment-service/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	db := repository.NewDB("postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable")

	repo := repository.NewPaymentRepo(db)
	usecase := usecase.NewPaymentUsecase(repo)
	handler := http.NewPaymentHandler(usecase)

	r := gin.Default()

	r.POST("/payments", handler.CreatePayment)
	r.GET("/payments/:order_id", handler.GetPayment)

	r.Run(":8081")
}
