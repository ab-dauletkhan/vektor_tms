package memory

import (
	"context"
	"errors"
	"sync"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
	"github.com/ab-dauletkhan/vektor_tms/internal/ports"
)

var errNilShipment = errors.New("nil shipment")

type Repository struct {
	mu                    sync.RWMutex
	shipmentsByID         map[string]*domainshipment.Shipment
	shipmentIDByReference map[string]string
}

func New() *Repository {
	return &Repository{
		shipmentsByID:         make(map[string]*domainshipment.Shipment),
		shipmentIDByReference: make(map[string]string),
	}
}

func (r *Repository) Create(ctx context.Context, shipment *domainshipment.Shipment) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if shipment == nil {
		return errNilShipment
	}

	cloned, err := cloneShipment(shipment)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shipmentsByID[cloned.ID()]; exists {
		return ports.ErrDuplicateShipmentID
	}
	if _, exists := r.shipmentIDByReference[cloned.ReferenceNumber()]; exists {
		return ports.ErrDuplicateReference
	}

	r.shipmentsByID[cloned.ID()] = cloned
	r.shipmentIDByReference[cloned.ReferenceNumber()] = cloned.ID()
	return nil
}

func (r *Repository) GetByReference(ctx context.Context, referenceNumber string) (*domainshipment.Shipment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.shipmentIDByReference[referenceNumber]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	shipment, ok := r.shipmentsByID[id]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	return cloneShipment(shipment)
}

func (r *Repository) UpdateByReference(ctx context.Context, referenceNumber string, updateFn ports.ShipmentUpdateFn) (*domainshipment.Shipment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.shipmentIDByReference[referenceNumber]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	storedShipment, ok := r.shipmentsByID[id]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	workingCopy, err := cloneShipment(storedShipment)
	if err != nil {
		return nil, err
	}

	if err := updateFn(workingCopy); err != nil {
		return nil, err
	}

	if existingID, exists := r.shipmentIDByReference[workingCopy.ReferenceNumber()]; exists && existingID != id {
		return nil, ports.ErrDuplicateReference
	}

	persistedCopy, err := cloneShipment(workingCopy)
	if err != nil {
		return nil, err
	}

	previousReference := storedShipment.ReferenceNumber()
	r.shipmentsByID[id] = persistedCopy
	if previousReference != persistedCopy.ReferenceNumber() {
		delete(r.shipmentIDByReference, previousReference)
	}
	r.shipmentIDByReference[persistedCopy.ReferenceNumber()] = id

	return cloneShipment(persistedCopy)
}

func cloneShipment(shipment *domainshipment.Shipment) (*domainshipment.Shipment, error) {
	if shipment == nil {
		return nil, errNilShipment
	}

	return shipment.Clone(), nil
}
