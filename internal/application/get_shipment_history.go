package application

import "context"

func (s *Service) GetShipmentHistory(ctx context.Context, query GetShipmentHistoryQuery) (*ShipmentHistory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	referenceNumber, err := normalizeReferenceNumber(query.ReferenceNumber)
	if err != nil {
		return nil, err
	}

	shipment, err := s.repository.GetByReference(ctx, referenceNumber)
	if err != nil {
		return nil, err
	}

	events := shipment.Events()
	history := make([]ShipmentEvent, len(events))
	for i, event := range events {
		history[i] = toShipmentEventDTO(event)
	}

	return &ShipmentHistory{
		ShipmentID:      shipment.ID(),
		ReferenceNumber: shipment.ReferenceNumber(),
		Events:          history,
	}, nil
}
