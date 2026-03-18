package shipment

import "errors"

var (
	ErrInvalidShipmentID       = errors.New("invalid shipment id")
	ErrInvalidReferenceNumber  = errors.New("invalid reference number")
	ErrInvalidLocation         = errors.New("invalid location")
	ErrOriginEqualsDestination = errors.New("origin and destination must differ")

	ErrInvalidDriver = errors.New("invalid driver")
	ErrInvalidUnit   = errors.New("invalid unit")
	ErrInvalidMoney  = errors.New("invalid money")

	ErrInvalidCreatedAt     = errors.New("invalid created_at")
	ErrInvalidOccurredAt    = errors.New("invalid occurred_at")
	ErrInvalidEventSequence = errors.New("invalid event sequence")
	ErrEmptyEventHistory    = errors.New("shipment history cannot be empty")
	ErrEventOutOfOrder      = errors.New("shipment event is out of chronological order")

	ErrRevenueExceedsAmount = errors.New("driver revenue cannot exceed shipment amount")
	ErrInvalidStatus        = errors.New("invalid shipment status")
	ErrInvalidInitialStatus = errors.New("initial shipment status must be pending")
	ErrInvalidTransition    = errors.New("invalid shipment status transition")
	ErrDuplicateStatus      = errors.New("duplicate shipment status")
)
