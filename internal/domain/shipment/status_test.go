package shipment

import (
	"errors"
	"testing"
)

func TestStatusIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "pending is valid",
			status: StatusPending,
			want:   true,
		},
		{
			name:   "picked_up is valid",
			status: StatusPickedUp,
			want:   true,
		},
		{
			name:   "in_transit is valid",
			status: StatusInTransit,
			want:   true,
		},
		{
			name:   "delivered is valid",
			status: StatusDelivered,
			want:   true,
		},
		{
			name:   "cancelled is valid",
			status: StatusCancelled,
			want:   true,
		},
		{
			name:   "empty status is invalid",
			status: "",
			want:   false,
		},
		{
			name:   "assigned status is invalid",
			status: Status("assigned"),
			want:   false,
		},
		{
			name:   "upper case pending is invalid",
			status: Status("PENDING"),
			want:   false,
		},
		{
			name:   "whitespace wrapped pending is invalid",
			status: Status(" pending "),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Fatalf("Status(%q).IsValid() = %t, want %t", tt.status, got, tt.want)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current Status
		next    Status
		wantErr error
	}{
		{
			name:    "pending to picked_up",
			current: StatusPending,
			next:    StatusPickedUp,
		},
		{
			name:    "pending to cancelled",
			current: StatusPending,
			next:    StatusCancelled,
		},
		{
			name:    "picked_up to in_transit",
			current: StatusPickedUp,
			next:    StatusInTransit,
		},
		{
			name:    "in_transit to delivered",
			current: StatusInTransit,
			next:    StatusDelivered,
		},
		{
			name:    "invalid current status",
			current: Status("assigned"),
			next:    StatusPickedUp,
			wantErr: ErrInvalidStatus,
		},
		{
			name:    "invalid next status",
			current: StatusPending,
			next:    Status("assigned"),
			wantErr: ErrInvalidStatus,
		},
		{
			name:    "duplicate status",
			current: StatusPending,
			next:    StatusPending,
			wantErr: ErrDuplicateStatus,
		},
		{
			name:    "pending to in_transit is invalid",
			current: StatusPending,
			next:    StatusInTransit,
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "picked_up to cancelled is invalid",
			current: StatusPickedUp,
			next:    StatusCancelled,
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "delivered is terminal",
			current: StatusDelivered,
			next:    StatusCancelled,
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "cancelled is terminal",
			current: StatusCancelled,
			next:    StatusPickedUp,
			wantErr: ErrInvalidTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.current, tt.next)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateTransition(%q, %q) error = %v, want %v", tt.current, tt.next, err, tt.wantErr)
			}
		})
	}
}
