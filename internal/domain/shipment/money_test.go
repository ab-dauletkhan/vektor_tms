package shipment

import (
	"errors"
	"testing"
)

func TestNewMoney(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		minorUnits int64
		wantMinor  int64
		wantErr    error
	}{
		{
			name:       "zero is allowed",
			minorUnits: 0,
			wantMinor:  0,
		},
		{
			name:       "positive amount is allowed",
			minorUnits: 1250,
			wantMinor:  1250,
		},
		{
			name:       "negative amount is rejected",
			minorUnits: -1,
			wantErr:    ErrInvalidMoney,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoney(tt.minorUnits)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewMoney() error = %v, want %v", err, tt.wantErr)
			}
			if money.MinorUnits() != tt.wantMinor {
				t.Fatalf("NewMoney().MinorUnits() = %d, want %d", money.MinorUnits(), tt.wantMinor)
			}
			if tt.wantErr == nil && !money.IsValid() {
				t.Fatal("NewMoney() returned invalid money for a valid amount")
			}
		})
	}
}
