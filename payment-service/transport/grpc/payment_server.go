package grpc

import (
	"context"
	"errors"
	"payment-service/domain"
	"payment-service/usecase"

	paymentv1 "github.com/almanac13/ADP2_asik2_generated/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	usecase *usecase.PaymentUsecase
}

func NewPaymentServer(u *usecase.PaymentUsecase) *PaymentServer {
	return &PaymentServer{usecase: u}
}

func toPaymentResponse(payment domain.Payment) *paymentv1.PaymentResponse {
	return &paymentv1.PaymentResponse{
		PaymentId:     payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     timestamppb.New(payment.CreatedAt),
	}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	payment, err := s.usecase.ProcessPayment(req.OrderId, req.Amount, req.IdempotencyKey)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidAmount) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toPaymentResponse(*payment), nil
}

func (s *PaymentServer) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, err := s.usecase.ListPayments(req.Status)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp := &paymentv1.ListPaymentsResponse{
		Payments: make([]*paymentv1.PaymentResponse, 0, len(payments)),
	}

	for _, payment := range payments {
		resp.Payments = append(resp.Payments, toPaymentResponse(payment))
	}

	return resp, nil
}
