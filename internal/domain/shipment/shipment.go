package shipment

import (
	"strings"
	"time"
)

type Shipment struct {
	id              string
	referenceNumber string
	origin          string
	destination     string
	driver          Driver
	unit            Unit
	shipmentAmount  Money
	driverRevenue   Money
	events          []Event
}

type NewParams struct {
	ID              string
	ReferenceNumber string
	Origin          string
	Destination     string
	Driver          Driver
	Unit            Unit
	ShipmentAmount  *Money
	DriverRevenue   *Money
	CreatedAt       time.Time
}

type RehydrateParams struct {
	ID              string
	ReferenceNumber string
	Origin          string
	Destination     string
	Driver          Driver
	Unit            Unit
	ShipmentAmount  *Money
	DriverRevenue   *Money
	Events          []Event
}

func NewShipment(params NewParams) (*Shipment, error) {
	if params.CreatedAt.IsZero() {
		return nil, ErrInvalidCreatedAt
	}

	id, referenceNumber, origin, destination, err := normalizeShipmentMetadata(
		params.ID,
		params.ReferenceNumber,
		params.Origin,
		params.Destination,
	)
	if err != nil {
		return nil, err
	}

	driver, unit, err := normalizeShipmentActors(params.Driver, params.Unit)
	if err != nil {
		return nil, err
	}

	shipmentAmount, driverRevenue, err := normalizeMoneyValues(params.ShipmentAmount, params.DriverRevenue)
	if err != nil {
		return nil, err
	}

	initialEvent, err := RehydrateEvent(1, StatusPending, params.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &Shipment{
		id:              id,
		referenceNumber: referenceNumber,
		origin:          origin,
		destination:     destination,
		driver:          driver,
		unit:            unit,
		shipmentAmount:  shipmentAmount,
		driverRevenue:   driverRevenue,
		events:          []Event{initialEvent},
	}, nil
}

func Rehydrate(params RehydrateParams) (*Shipment, error) {
	id, referenceNumber, origin, destination, err := normalizeShipmentMetadata(
		params.ID,
		params.ReferenceNumber,
		params.Origin,
		params.Destination,
	)
	if err != nil {
		return nil, err
	}

	driver, unit, err := normalizeShipmentActors(params.Driver, params.Unit)
	if err != nil {
		return nil, err
	}

	shipmentAmount, driverRevenue, err := normalizeMoneyValues(params.ShipmentAmount, params.DriverRevenue)
	if err != nil {
		return nil, err
	}

	events, err := validateHistory(params.Events)
	if err != nil {
		return nil, err
	}

	return &Shipment{
		id:              id,
		referenceNumber: referenceNumber,
		origin:          origin,
		destination:     destination,
		driver:          driver,
		unit:            unit,
		shipmentAmount:  shipmentAmount,
		driverRevenue:   driverRevenue,
		events:          events,
	}, nil
}

func (s *Shipment) AddStatusEvent(next Status, occurredAt time.Time) (Event, error) {
	if occurredAt.IsZero() {
		return Event{}, ErrInvalidOccurredAt
	}

	last := s.lastEvent()
	if err := ValidateTransition(last.Status(), next); err != nil {
		return Event{}, err
	}

	occurredAt = occurredAt.UTC()
	if occurredAt.Before(last.OccurredAt()) {
		return Event{}, ErrEventOutOfOrder
	}

	event, err := RehydrateEvent(last.Sequence()+1, next, occurredAt)
	if err != nil {
		return Event{}, err
	}

	s.events = append(s.events, event)

	return event, nil
}

func (s *Shipment) ID() string {
	return s.id
}

func (s *Shipment) ReferenceNumber() string {
	return s.referenceNumber
}

func (s *Shipment) Origin() string {
	return s.origin
}

func (s *Shipment) Destination() string {
	return s.destination
}

func (s *Shipment) Driver() Driver {
	return s.driver
}

func (s *Shipment) Unit() Unit {
	return s.unit
}

func (s *Shipment) ShipmentAmount() Money {
	return s.shipmentAmount
}

func (s *Shipment) DriverRevenue() Money {
	return s.driverRevenue
}

func (s *Shipment) CreatedAt() time.Time {
	return s.events[0].OccurredAt()
}

func (s *Shipment) UpdatedAt() time.Time {
	return s.lastEvent().OccurredAt()
}

func (s *Shipment) CurrentStatus() Status {
	return s.lastEvent().Status()
}

func (s *Shipment) Events() []Event {
	copied := make([]Event, len(s.events))
	copy(copied, s.events)
	return copied
}

func (s *Shipment) lastEvent() Event {
	return s.events[len(s.events)-1]
}

func normalizeShipmentMetadata(
	id string,
	referenceNumber string,
	origin string,
	destination string,
) (string, string, string, string, error) {
	id = strings.TrimSpace(id)
	referenceNumber = strings.TrimSpace(referenceNumber)
	origin = strings.TrimSpace(origin)
	destination = strings.TrimSpace(destination)

	if id == "" {
		return "", "", "", "", ErrInvalidShipmentID
	}
	if referenceNumber == "" {
		return "", "", "", "", ErrInvalidReferenceNumber
	}
	if origin == "" || destination == "" {
		return "", "", "", "", ErrInvalidLocation
	}
	// Locations are treated as opaque user-provided labels, so equality is
	// literal after trimming.
	if origin == destination {
		return "", "", "", "", ErrOriginEqualsDestination
	}

	return id, referenceNumber, origin, destination, nil
}

func normalizeShipmentActors(driver Driver, unit Unit) (Driver, Unit, error) {
	normalizedDriver, err := NewDriver(driver.ID(), driver.Name())
	if err != nil {
		return Driver{}, Unit{}, err
	}

	normalizedUnit, err := NewUnit(unit.ID(), unit.RegistrationNumber())
	if err != nil {
		return Driver{}, Unit{}, err
	}

	return normalizedDriver, normalizedUnit, nil
}

func normalizeMoneyValues(shipmentAmount, driverRevenue *Money) (Money, Money, error) {
	if shipmentAmount == nil || driverRevenue == nil {
		return Money{}, Money{}, ErrInvalidMoney
	}
	if !shipmentAmount.IsValid() || !driverRevenue.IsValid() {
		return Money{}, Money{}, ErrInvalidMoney
	}
	if driverRevenue.GreaterThan(*shipmentAmount) {
		return Money{}, Money{}, ErrRevenueExceedsAmount
	}

	return *shipmentAmount, *driverRevenue, nil
}

func validateHistory(history []Event) ([]Event, error) {
	if len(history) == 0 {
		return nil, ErrEmptyEventHistory
	}

	validated := make([]Event, len(history))
	for i, event := range history {
		normalized, err := RehydrateEvent(event.sequence, event.status, event.occurredAt)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			if normalized.Sequence() != 1 {
				return nil, ErrInvalidEventSequence
			}
			if normalized.Status() != StatusPending {
				return nil, ErrInvalidInitialStatus
			}

			validated[i] = normalized
			continue
		}

		previous := validated[i-1]
		if normalized.Sequence() != previous.Sequence()+1 {
			return nil, ErrInvalidEventSequence
		}
		if normalized.OccurredAt().Before(previous.OccurredAt()) {
			return nil, ErrEventOutOfOrder
		}
		if err := ValidateTransition(previous.Status(), normalized.Status()); err != nil {
			return nil, err
		}

		validated[i] = normalized
	}

	return validated, nil
}
