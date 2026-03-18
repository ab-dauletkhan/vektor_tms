package grpcadapter

import (
	"context"
	"io"
	"log"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestRecoveryUnaryInterceptorReturnsInternalOnPanic(t *testing.T) {
	t.Parallel()

	interceptor := recoveryUnaryInterceptor(log.New(io.Discard, "", 0))
	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/shipment.v1.ShipmentService/GetShipment"},
		func(context.Context, any) (any, error) {
			panic("boom")
		},
	)
	if got := grpcstatus.Code(err); got != codes.Internal {
		t.Fatalf("recoveryUnaryInterceptor() code = %v, want %v", got, codes.Internal)
	}
}
