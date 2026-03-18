package application

import "context"

func (s *Service) GetShipment(ctx context.Context, query GetShipmentQuery) (*Shipment, error) {
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

	response := toShipmentDTO(shipment)
	return &response, nil
}
