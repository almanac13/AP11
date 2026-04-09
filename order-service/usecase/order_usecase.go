package usecase

import (
	"errors"
	"order-service/domain"
	"time"

	"github.com/google/uuid"
)

type OrderUsecase struct {
	repo    OrderRepository
	payment *PaymentClient
}

func NewOrderUsecase(r OrderRepository, p *PaymentClient) *OrderUsecase {
	return &OrderUsecase{repo: r, payment: p}
}

func (u *OrderUsecase) GetOrder(id string) (*domain.Order, error) {
	return u.repo.GetByID(id)
}

func (u *OrderUsecase) CancelOrder(id string) (*domain.Order, error) {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.Status != "Pending" {
		return nil, errors.New("only pending orders can be cancelled")
	}

	err = u.repo.UpdateStatus(id, "Cancelled")
	if err != nil {
		return nil, err
	}

	order.Status = "Cancelled"
	return order, nil
}

func (u *OrderUsecase) CreateOrder(customerID, itemName string, amount int64) (*domain.Order, error) {

	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	order := domain.Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     "Pending",
		CreatedAt:  time.Now(),
	}

	// Save first
	err := u.repo.Create(order)
	if err != nil {
		return nil, err
	}

	// Call payment service
	status, err := u.payment.Pay(order.ID, amount)
	if err != nil {
		u.repo.UpdateStatus(order.ID, "Failed")
		return nil, err
	}

	if status == "Authorized" {
		u.repo.UpdateStatus(order.ID, "Paid")
		order.Status = "Paid"
	} else {
		u.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
	}

	return &order, nil
}

func (u *OrderUsecase) GetOrdersByCustomer(customerID string) ([]domain.Order, int, error) {
	if customerID == "" {
		return nil, 0, errors.New("customer_id is required")
	}

	orders, err := u.repo.GetByCustomerID(customerID)
	if err != nil {
		return nil, 0, err
	}

	if len(orders) == 0 {
		return nil, 0, errors.New("no orders found for this customer")
	}

	return orders, len(orders), nil

}
