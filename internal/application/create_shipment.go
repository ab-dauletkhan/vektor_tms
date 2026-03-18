package application

import (
	"context"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
)

func (s *Service) CreateShipment(ctx context.Context, command CreateShipmentCommand) (*Shipment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	driver, err := toDomainDriver(command.Driver)
	if err != nil {
		return nil, err
	}

	unit, err := toDomainUnit(command.Unit)
	if err != nil {
		return nil, err
	}

	shipmentAmount, err := toDomainMoney(command.ShipmentAmountMinor)
	if err != nil {
		return nil, err
	}

	driverRevenue, err := toDomainMoney(command.DriverRevenueMinor)
	if err != nil {
		return nil, err
	}

	shipment, err := domainshipment.NewShipment(domainshipment.NewParams{
		ID:              s.idGenerator.NewID(),
		ReferenceNumber: command.ReferenceNumber,
		Origin:          command.Origin,
		Destination:     command.Destination,
		Driver:          driver,
		Unit:            unit,
		ShipmentAmount:  shipmentAmount,
		DriverRevenue:   driverRevenue,
		CreatedAt:       s.clock.Now(),
	})
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, shipment); err != nil {
		return nil, err
	}

	response := toShipmentDTO(shipment)
	return &response, nil
}
