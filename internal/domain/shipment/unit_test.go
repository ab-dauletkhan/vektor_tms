package shipment

import (
	"errors"
	"testing"
)

func TestNewUnit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		id           string
		registration string
		wantID       string
		wantReg      string
		wantErr      error
	}{
		{
			name:         "trims valid fields",
			id:           " unit-1 ",
			registration: " AB123CD ",
			wantID:       "unit-1",
			wantReg:      "AB123CD",
		},
		{
			name:         "rejects blank id",
			id:           " ",
			registration: "AB123CD",
			wantErr:      ErrInvalidUnit,
		},
		{
			name:         "rejects blank registration number",
			id:           "unit-1",
			registration: "\t",
			wantErr:      ErrInvalidUnit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit, err := NewUnit(tt.id, tt.registration)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewUnit() error = %v, want %v", err, tt.wantErr)
			}
			if unit.ID() != tt.wantID || unit.RegistrationNumber() != tt.wantReg {
				t.Fatalf("NewUnit() = {id:%q registration:%q}, want {id:%q registration:%q}", unit.ID(), unit.RegistrationNumber(), tt.wantID, tt.wantReg)
			}
		})
	}
}
