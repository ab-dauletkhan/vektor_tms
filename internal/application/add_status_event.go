package application

import (
	"context"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
)

func (s *Service) AddStatusEvent(ctx context.Context, command AddShipmentStatusEventCommand) (*AddStatusEventResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	referenceNumber, err := normalizeReferenceNumber(command.ReferenceNumber)
	if err != nil {
		return nil, err
	}

	status, err := toDomainStatus(command.Status)
	if err != nil {
		return nil, err
	}

	occurredAt := s.clock.Now()
	var createdEvent ShipmentEvent

	shipment, err := s.repository.UpdateByReference(ctx, referenceNumber, func(shipment *domainshipment.Shipment) error {
		event, err := shipment.AddStatusEvent(status, occurredAt)
		if err != nil {
			return err
		}

		createdEvent = toShipmentEventDTO(event)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AddStatusEventResult{
		Shipment: toShipmentDTO(shipment),
		Event:    createdEvent,
	}, nil
}
