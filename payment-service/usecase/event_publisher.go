package usecase

import "payment-service/domain"

type EventPublisher interface {
	PublishPaymentCompleted(payment domain.Payment) error
}
