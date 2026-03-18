package shipment

import "slices"

type Status string

const (
	StatusPending   Status = "pending"
	StatusPickedUp  Status = "picked_up"
	StatusInTransit Status = "in_transit"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

// Assumption: once a shipment is picked up it must continue through transit and
// cannot be cancelled anymore.
var allowedTransitions = map[Status][]Status{
	StatusPending:   {StatusPickedUp, StatusCancelled},
	StatusPickedUp:  {StatusInTransit},
	StatusInTransit: {StatusDelivered},
	StatusDelivered: nil,
	StatusCancelled: nil,
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusPickedUp, StatusInTransit, StatusDelivered, StatusCancelled:
		return true
	default:
		return false
	}
}

func ValidateTransition(current, next Status) error {
	if !current.IsValid() || !next.IsValid() {
		return ErrInvalidStatus
	}
	if current == next {
		return ErrDuplicateStatus
	}
	if !current.canTransitionTo(next) {
		return ErrInvalidTransition
	}

	return nil
}

func (s Status) canTransitionTo(next Status) bool {
	if !s.IsValid() || !next.IsValid() {
		return false
	}

	return slices.Contains(allowedTransitions[s], next)
}
