package grpcadapter

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
	"github.com/ab-dauletkhan/vektor_tms/internal/ports"
)

func TestToGRPCError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{
			name: "invalid argument",
			err:  domainshipment.ErrInvalidMoney,
			want: codes.InvalidArgument,
		},
		{
			name: "nil request",
			err:  errNilRequest,
			want: codes.InvalidArgument,
		},
		{
			name: "not found",
			err:  ports.ErrShipmentNotFound,
			want: codes.NotFound,
		},
		{
			name: "duplicate reference",
			err:  ports.ErrDuplicateReference,
			want: codes.AlreadyExists,
		},
		{
			name: "duplicate shipment id",
			err:  ports.ErrDuplicateShipmentID,
			want: codes.AlreadyExists,
		},
		{
			name: "canceled",
			err:  context.Canceled,
			want: codes.Canceled,
		},
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
			want: codes.DeadlineExceeded,
		},
		{
			name: "internal",
			err:  errors.New("boom"),
			want: codes.Internal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := grpcstatus.Code(toGRPCError(tt.err)); got != tt.want {
				t.Fatalf("toGRPCError() code = %v, want %v", got, tt.want)
			}
		})
	}
}
