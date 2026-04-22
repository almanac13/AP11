package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	log.Printf("method=%s duration=%s error=%v", info.FullMethod, time.Since(start), err)

	return resp, err
}
