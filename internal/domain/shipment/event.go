package shipment

import "time"

type Event struct {
	status     Status
	sequence   uint32
	occurredAt time.Time
}

func RehydrateEvent(sequence uint32, status Status, occurredAt time.Time) (Event, error) {
	if sequence == 0 {
		return Event{}, ErrInvalidEventSequence
	}
	if !status.IsValid() {
		return Event{}, ErrInvalidStatus
	}
	if occurredAt.IsZero() {
		return Event{}, ErrInvalidOccurredAt
	}

	return Event{
		status:     status,
		sequence:   sequence,
		occurredAt: occurredAt.UTC(),
	}, nil
}

func (e Event) Status() Status {
	return e.status
}

func (e Event) Sequence() uint32 {
	return e.sequence
}

func (e Event) OccurredAt() time.Time {
	return e.occurredAt
}
