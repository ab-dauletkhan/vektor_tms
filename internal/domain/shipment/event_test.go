package shipment

import (
	"errors"
	"testing"
	"time"
)

func TestRehydrateEvent(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("ALMT", 5*60*60)
	occurredAt := time.Date(2026, time.March, 18, 15, 4, 5, 0, loc)

	tests := []struct {
		name       string
		sequence   uint32
		status     Status
		occurredAt time.Time
		wantErr    error
		wantStatus Status
		wantSeq    uint32
		wantTime   time.Time
	}{
		{
			name:       "valid event is normalized to utc",
			sequence:   2,
			status:     StatusInTransit,
			occurredAt: occurredAt,
			wantStatus: StatusInTransit,
			wantSeq:    2,
			wantTime:   occurredAt.UTC(),
		},
		{
			name:       "zero sequence is rejected",
			sequence:   0,
			status:     StatusPending,
			occurredAt: occurredAt,
			wantErr:    ErrInvalidEventSequence,
		},
		{
			name:       "invalid status is rejected",
			sequence:   1,
			status:     Status("assigned"),
			occurredAt: occurredAt,
			wantErr:    ErrInvalidStatus,
		},
		{
			name:       "zero time is rejected",
			sequence:   1,
			status:     StatusPending,
			occurredAt: time.Time{},
			wantErr:    ErrInvalidOccurredAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := RehydrateEvent(tt.sequence, tt.status, tt.occurredAt)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RehydrateEvent() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if event != (Event{}) {
					t.Fatalf("RehydrateEvent() = %#v, want zero value", event)
				}
				return
			}

			if got := event.Sequence(); got != tt.wantSeq {
				t.Fatalf("event.Sequence() = %d, want %d", got, tt.wantSeq)
			}
			if got := event.Status(); got != tt.wantStatus {
				t.Fatalf("event.Status() = %q, want %q", got, tt.wantStatus)
			}
			if got := event.OccurredAt(); !got.Equal(tt.wantTime) {
				t.Fatalf("event.OccurredAt() = %s, want %s", got, tt.wantTime)
			}
			if got := event.OccurredAt().Location(); got != time.UTC {
				t.Fatalf("event.OccurredAt() location = %v, want UTC", got)
			}
		})
	}
}
