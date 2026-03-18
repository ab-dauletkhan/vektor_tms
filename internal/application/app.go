package application

import (
	"errors"

	"github.com/ab-dauletkhan/vektor_tms/internal/ports"
)

var (
	ErrNilShipmentRepository = errors.New("nil shipment repository")
	ErrNilClock              = errors.New("nil clock")
	ErrNilIDGenerator        = errors.New("nil id generator")
)

type Service struct {
	repository  ports.ShipmentRepository
	clock       ports.Clock
	idGenerator ports.IDGenerator
}

func New(repository ports.ShipmentRepository, clock ports.Clock, idGenerator ports.IDGenerator) (*Service, error) {
	if repository == nil {
		return nil, ErrNilShipmentRepository
	}
	if clock == nil {
		return nil, ErrNilClock
	}
	if idGenerator == nil {
		return nil, ErrNilIDGenerator
	}

	return &Service{
		repository:  repository,
		clock:       clock,
		idGenerator: idGenerator,
	}, nil
}
