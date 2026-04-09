package usecase

import (
	"errors"
	"payment-service/domain"

	"github.com/google/uuid"
)

type PaymentUsecase struct {
	repo PaymentRepository
}

func NewPaymentUsecase(r PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{repo: r}
}

func (u *PaymentUsecase) GetPayment(orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(orderID)
}

func (u *PaymentUsecase) ProcessPayment(orderID string, amount int64, idempotencyKey string) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	// Idempotency: if payment for this order already exists, return it (BONUS)
	if idempotencyKey != "" {
		existing, err := u.repo.GetByOrderID(orderID)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	payment := domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
	}

	// BUSINESS RULE: amount > 100000 → Declined
	if amount > 100000 {
		payment.Status = "Declined"
	} else {
		payment.Status = "Authorized"
	}

	err := u.repo.Create(payment)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}
