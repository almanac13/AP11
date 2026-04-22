package grpc

import (
	"io"
	"order-service/repository"

	ordertrackingv1 "github.com/almanac13/ADP2_asik2_generated/ordertracking/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderUpdatesRepository interface {
	Subscribe(orderID string) chan repository.OrderStatusEvent
	Unsubscribe(orderID string, target chan repository.OrderStatusEvent)
}

type OrderTrackingServer struct {
	ordertrackingv1.UnimplementedOrderTrackingServiceServer
	repo OrderUpdatesRepository
}

func NewOrderTrackingServer(repo OrderUpdatesRepository) *OrderTrackingServer {
	return &OrderTrackingServer{repo: repo}
}

func (s *OrderTrackingServer) SubscribeToOrderUpdates(
	req *ordertrackingv1.OrderRequest,
	stream ordertrackingv1.OrderTrackingService_SubscribeToOrderUpdatesServer,
) error {
	ch := s.repo.Subscribe(req.OrderId)
	defer s.repo.Unsubscribe(req.OrderId, ch)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event, ok := <-ch:
			if !ok {
				return io.EOF
			}
			if err := stream.Send(&ordertrackingv1.OrderStatusUpdate{
				OrderId:   event.OrderID,
				Status:    event.Status,
				UpdatedAt: timestamppb.New(event.UpdatedAt),
			}); err != nil {
				return err
			}
		}
	}
}
