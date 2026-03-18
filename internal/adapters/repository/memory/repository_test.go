package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
	"github.com/ab-dauletkhan/vektor_tms/internal/ports"
)

func TestRepositoryCreateAndGetByReferenceReturnDetachedCopies(t *testing.T) {
	t.Parallel()

	repository := New()
	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	shipment := mustShipment(t, "shipment-1", "REF-001", createdAt)
	if err := repository.Create(context.Background(), shipment); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	loaded, err := repository.GetByReference(context.Background(), "REF-001")
	if err != nil {
		t.Fatalf("GetByReference() error = %v", err)
	}

	if _, err := loaded.AddStatusEvent(domainshipment.StatusPickedUp, createdAt.Add(time.Hour)); err != nil {
		t.Fatalf("loaded.AddStatusEvent() error = %v", err)
	}

	reloaded, err := repository.GetByReference(context.Background(), "REF-001")
	if err != nil {
		t.Fatalf("GetByReference() second call error = %v", err)
	}

	if got := reloaded.CurrentStatus(); got != domainshipment.StatusPending {
		t.Fatalf("Reloaded CurrentStatus() = %q, want %q", got, domainshipment.StatusPending)
	}
}

func TestRepositoryCreateRejectsDuplicateReference(t *testing.T) {
	t.Parallel()

	repository := New()
	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	if err := repository.Create(context.Background(), mustShipment(t, "shipment-1", "REF-001", createdAt)); err != nil {
		t.Fatalf("Create() first call error = %v", err)
	}

	err := repository.Create(context.Background(), mustShipment(t, "shipment-2", "REF-001", createdAt.Add(time.Minute)))
	if !errors.Is(err, ports.ErrDuplicateReference) {
		t.Fatalf("Create() second call error = %v, want %v", err, ports.ErrDuplicateReference)
	}
}

func TestRepositoryCreateRejectsDuplicateID(t *testing.T) {
	t.Parallel()

	repository := New()
	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	if err := repository.Create(context.Background(), mustShipment(t, "shipment-1", "REF-001", createdAt)); err != nil {
		t.Fatalf("Create() first call error = %v", err)
	}

	err := repository.Create(context.Background(), mustShipment(t, "shipment-1", "REF-002", createdAt.Add(time.Minute)))
	if !errors.Is(err, ports.ErrDuplicateShipmentID) {
		t.Fatalf("Create() second call error = %v, want %v", err, ports.ErrDuplicateShipmentID)
	}
}

func TestRepositoryUpdateByReferencePersistsMutation(t *testing.T) {
	t.Parallel()

	repository := New()
	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	if err := repository.Create(context.Background(), mustShipment(t, "shipment-1", "REF-001", createdAt)); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repository.UpdateByReference(context.Background(), "REF-001", func(shipment *domainshipment.Shipment) error {
		_, err := shipment.AddStatusEvent(domainshipment.StatusPickedUp, createdAt.Add(time.Hour))
		return err
	})
	if err != nil {
		t.Fatalf("UpdateByReference() error = %v", err)
	}

	if got := updated.CurrentStatus(); got != domainshipment.StatusPickedUp {
		t.Fatalf("Updated CurrentStatus() = %q, want %q", got, domainshipment.StatusPickedUp)
	}

	reloaded, err := repository.GetByReference(context.Background(), "REF-001")
	if err != nil {
		t.Fatalf("GetByReference() error = %v", err)
	}
	if got := reloaded.CurrentStatus(); got != domainshipment.StatusPickedUp {
		t.Fatalf("Reloaded CurrentStatus() = %q, want %q", got, domainshipment.StatusPickedUp)
	}
}

func TestRepositoryUpdateByReferenceRollsBackOnError(t *testing.T) {
	t.Parallel()

	repository := New()
	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	if err := repository.Create(context.Background(), mustShipment(t, "shipment-1", "REF-001", createdAt)); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updateErr := errors.New("update failed after mutation")
	_, err := repository.UpdateByReference(context.Background(), "REF-001", func(shipment *domainshipment.Shipment) error {
		if _, err := shipment.AddStatusEvent(domainshipment.StatusPickedUp, createdAt.Add(time.Hour)); err != nil {
			return err
		}

		return updateErr
	})
	if !errors.Is(err, updateErr) {
		t.Fatalf("UpdateByReference() error = %v, want %v", err, updateErr)
	}

	reloaded, err := repository.GetByReference(context.Background(), "REF-001")
	if err != nil {
		t.Fatalf("GetByReference() error = %v", err)
	}

	if got := reloaded.CurrentStatus(); got != domainshipment.StatusPending {
		t.Fatalf("Reloaded CurrentStatus() = %q, want %q", got, domainshipment.StatusPending)
	}
	if got := len(reloaded.Events()); got != 1 {
		t.Fatalf("len(Reloaded.Events()) = %d, want 1", got)
	}
}

func TestRepositoryGetByReferenceReturnsNotFound(t *testing.T) {
	t.Parallel()

	repository := New()
	_, err := repository.GetByReference(context.Background(), "REF-404")

	if !errors.Is(err, ports.ErrShipmentNotFound) {
		t.Fatalf("GetByReference() error = %v, want %v", err, ports.ErrShipmentNotFound)
	}
}

func TestRepositoryUpdateByReferenceReturnsNotFound(t *testing.T) {
	t.Parallel()

	repository := New()
	_, err := repository.UpdateByReference(context.Background(), "REF-404", func(shipment *domainshipment.Shipment) error {
		return nil
	})

	if !errors.Is(err, ports.ErrShipmentNotFound) {
		t.Fatalf("UpdateByReference() error = %v, want %v", err, ports.ErrShipmentNotFound)
	}
}

func mustShipment(t *testing.T, id, reference string, createdAt time.Time) *domainshipment.Shipment {
	t.Helper()

	driver, err := domainshipment.NewDriver("driver-1", "Jane Doe")
	if err != nil {
		t.Fatalf("NewDriver() error = %v", err)
	}

	unit, err := domainshipment.NewUnit("unit-1", "123ABC02")
	if err != nil {
		t.Fatalf("NewUnit() error = %v", err)
	}

	shipmentAmount, err := domainshipment.NewMoney(100_00)
	if err != nil {
		t.Fatalf("NewMoney(shipment amount) error = %v", err)
	}

	driverRevenue, err := domainshipment.NewMoney(60_00)
	if err != nil {
		t.Fatalf("NewMoney(driver revenue) error = %v", err)
	}

	shipment, err := domainshipment.NewShipment(domainshipment.NewParams{
		ID:              id,
		ReferenceNumber: reference,
		Origin:          "Almaty",
		Destination:     "Astana",
		Driver:          driver,
		Unit:            unit,
		ShipmentAmount:  &shipmentAmount,
		DriverRevenue:   &driverRevenue,
		CreatedAt:       createdAt,
	})
	if err != nil {
		t.Fatalf("NewShipment() error = %v", err)
	}

	return shipment
}
