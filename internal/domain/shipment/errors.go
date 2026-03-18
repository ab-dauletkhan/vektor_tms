package shipment

type invalidArgumentError string

func (e invalidArgumentError) Error() string {
	return string(e)
}

func (invalidArgumentError) IsInvalidArgument() bool {
	return true
}

var (
	ErrInvalidShipmentID       = invalidArgumentError("invalid shipment id")
	ErrInvalidReferenceNumber  = invalidArgumentError("invalid reference number")
	ErrInvalidLocation         = invalidArgumentError("invalid location")
	ErrOriginEqualsDestination = invalidArgumentError("origin and destination must differ")

	ErrInvalidDriver = invalidArgumentError("invalid driver")
	ErrInvalidUnit   = invalidArgumentError("invalid unit")
	ErrInvalidMoney  = invalidArgumentError("invalid money")

	ErrInvalidCreatedAt     = invalidArgumentError("invalid created_at")
	ErrInvalidOccurredAt    = invalidArgumentError("invalid occurred_at")
	ErrInvalidEventSequence = invalidArgumentError("invalid event sequence")
	ErrEmptyEventHistory    = invalidArgumentError("shipment history cannot be empty")
	ErrEventOutOfOrder      = invalidArgumentError("shipment event is out of chronological order")

	ErrRevenueExceedsAmount = invalidArgumentError("driver revenue cannot exceed shipment amount")
	ErrInvalidStatus        = invalidArgumentError("invalid shipment status")
	ErrInvalidInitialStatus = invalidArgumentError("initial shipment status must be pending")
	ErrInvalidTransition    = invalidArgumentError("invalid shipment status transition")
	ErrDuplicateStatus      = invalidArgumentError("duplicate shipment status")
)
