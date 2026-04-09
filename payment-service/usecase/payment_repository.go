package usecase

import "payment-service/domain"

type PaymentRepository interface {
	Create(payment domain.Payment) error
	GetByOrderID(orderID string) (*domain.Payment, error)
}
