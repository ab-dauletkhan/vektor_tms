package application

import (
	"strings"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
)

func toDomainDriver(input DriverInput) (domainshipment.Driver, error) {
	return domainshipment.NewDriver(input.ID, input.Name)
}

func toDomainUnit(input UnitInput) (domainshipment.Unit, error) {
	return domainshipment.NewUnit(input.ID, input.RegistrationNumber)
}

func toDomainMoney(minorUnits *int64) (*domainshipment.Money, error) {
	if minorUnits == nil {
		return nil, domainshipment.ErrInvalidMoney
	}

	money, err := domainshipment.NewMoney(*minorUnits)
	if err != nil {
		return nil, err
	}

	return &money, nil
}

func toDomainStatus(value string) (domainshipment.Status, error) {
	status := domainshipment.Status(value)
	if !status.IsValid() {
		return "", domainshipment.ErrInvalidStatus
	}

	return status, nil
}

func normalizeReferenceNumber(referenceNumber string) (string, error) {
	referenceNumber = strings.TrimSpace(referenceNumber)
	if referenceNumber == "" {
		return "", domainshipment.ErrInvalidReferenceNumber
	}

	return referenceNumber, nil
}

func toShipmentDTO(shipment *domainshipment.Shipment) Shipment {
	return Shipment{
		ID:              shipment.ID(),
		ReferenceNumber: shipment.ReferenceNumber(),
		Origin:          shipment.Origin(),
		Destination:     shipment.Destination(),
		CurrentStatus:   string(shipment.CurrentStatus()),
		Driver: Driver{
			ID:   shipment.Driver().ID(),
			Name: shipment.Driver().Name(),
		},
		Unit: Unit{
			ID:                 shipment.Unit().ID(),
			RegistrationNumber: shipment.Unit().RegistrationNumber(),
		},
		ShipmentAmountMinor: shipment.ShipmentAmount().MinorUnits(),
		DriverRevenueMinor:  shipment.DriverRevenue().MinorUnits(),
		CreatedAt:           shipment.CreatedAt(),
		UpdatedAt:           shipment.UpdatedAt(),
	}
}

func toShipmentEventDTO(event domainshipment.Event) ShipmentEvent {
	return ShipmentEvent{
		Sequence:   event.Sequence(),
		Status:     string(event.Status()),
		OccurredAt: event.OccurredAt(),
	}
}
