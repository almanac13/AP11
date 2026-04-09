package usecase

import "order-service/domain"

type OrderRepository interface {
	Create(order domain.Order) error
	GetByID(id string) (*domain.Order, error)
	UpdateStatus(id string, status string) error
	GetByCustomerID(customerID string) ([]domain.Order, error)
}
