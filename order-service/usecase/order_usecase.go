package usecase

import (
	"context"
	"errors"
	"log"
	"order-service/cache"
	"order-service/domain"
	"time"

	"github.com/google/uuid"
)

type OrderUsecase struct {
	repo       OrderRepository
	payment    *PaymentClient
	orderCache cache.OrderCache
}

func NewOrderUsecase(r OrderRepository, p *PaymentClient, orderCache cache.OrderCache) *OrderUsecase {
	return &OrderUsecase{repo: r, payment: p, orderCache: orderCache}
}

func (u *OrderUsecase) GetOrder(id string) (*domain.Order, error) {
	ctx := context.Background()

	if u.orderCache != nil {
		order, err := u.orderCache.GetOrder(ctx, id)
		if err == nil {
			return order, nil
		}
		if !cache.IsCacheMiss(err) {
			log.Printf("[Redis] cache read error for order_id=%s: %v", id, err)
		} else {
			log.Printf("[Redis] cache MISS for order_id=%s", id)
		}
	}

	order, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if u.orderCache != nil {
		if err := u.orderCache.SetOrder(ctx, order); err != nil {
			log.Printf("[Redis] cache write error for order_id=%s: %v", id, err)
		}
	}

	return order, nil
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

	u.invalidateOrderCache(id)
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
		u.invalidateOrderCache(order.ID)
		order.Status = "Paid"
	} else {
		u.repo.UpdateStatus(order.ID, "Failed")
		u.invalidateOrderCache(order.ID)
		order.Status = "Failed"
	}

	return &order, nil
}

func (u *OrderUsecase) invalidateOrderCache(id string) {
	if u.orderCache == nil {
		return
	}
	if err := u.orderCache.DeleteOrder(context.Background(), id); err != nil {
		log.Printf("[Redis] cache invalidation error for order_id=%s: %v", id, err)
	}
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
