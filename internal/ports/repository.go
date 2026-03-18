package ports

import (
	"context"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
)

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}

func (notFoundError) IsNotFound() bool {
	return true
}

type alreadyExistsError string

func (e alreadyExistsError) Error() string {
	return string(e)
}

func (alreadyExistsError) IsAlreadyExists() bool {
	return true
}

var (
	ErrShipmentNotFound    = notFoundError("shipment not found")
	ErrDuplicateShipmentID = alreadyExistsError("shipment id already exists")
	ErrDuplicateReference  = alreadyExistsError("shipment reference number already exists")
)

type ShipmentUpdateFn func(*domainshipment.Shipment) error

type ShipmentRepository interface {
	Create(ctx context.Context, shipment *domainshipment.Shipment) error
	GetByReference(ctx context.Context, referenceNumber string) (*domainshipment.Shipment, error)
	UpdateByReference(ctx context.Context, referenceNumber string, updateFn ShipmentUpdateFn) (*domainshipment.Shipment, error)
}
