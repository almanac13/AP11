package usecase

import (
	"errors"
	"payment-service/domain"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidAmount = errors.New("invalid amount")

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
		return nil, ErrInvalidAmount
	}

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
		CreatedAt:     time.Now(),
	}

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

func (u *PaymentUsecase) ListPayments(status string) ([]domain.Payment, error) {
	if status == "" {
		return nil, errors.New("status is required")
	}

	return u.repo.ListByStatus(status)
}
