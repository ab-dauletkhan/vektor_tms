package shipment

import (
	"errors"
	"testing"
)

func TestNewDriver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		fullName string
		wantID   string
		wantName string
		wantErr  error
	}{
		{
			name:     "trims valid fields",
			id:       " driver-1 ",
			fullName: " Aigerim ",
			wantID:   "driver-1",
			wantName: "Aigerim",
		},
		{
			name:     "rejects blank id",
			id:       "   ",
			fullName: "Aigerim",
			wantErr:  ErrInvalidDriver,
		},
		{
			name:     "rejects blank name",
			id:       "driver-1",
			fullName: "\t",
			wantErr:  ErrInvalidDriver,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := NewDriver(tt.id, tt.fullName)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewDriver() error = %v, want %v", err, tt.wantErr)
			}
			if driver.ID() != tt.wantID || driver.Name() != tt.wantName {
				t.Fatalf("NewDriver() = {id:%q name:%q}, want {id:%q name:%q}", driver.ID(), driver.Name(), tt.wantID, tt.wantName)
			}
		})
	}
}
