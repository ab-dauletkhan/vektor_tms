package application

import (
	"context"
	"errors"
	"testing"
	"time"

	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
	"github.com/ab-dauletkhan/vektor_tms/internal/ports"
)

func TestNewRejectsNilDependencies(t *testing.T) {
	t.Parallel()

	clock := stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)}
	idGenerator := stubIDGenerator{id: "shipment-1"}
	repository := newStubRepository()

	tests := []struct {
		name        string
		repository  ports.ShipmentRepository
		clock       ports.Clock
		idGenerator ports.IDGenerator
		wantErr     error
	}{
		{
			name:        "nil repository",
			clock:       clock,
			idGenerator: idGenerator,
			wantErr:     ErrNilShipmentRepository,
		},
		{
			name:        "nil clock",
			repository:  repository,
			idGenerator: idGenerator,
			wantErr:     ErrNilClock,
		},
		{
			name:       "nil id generator",
			repository: repository,
			clock:      clock,
			wantErr:    ErrNilIDGenerator,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(tt.repository, tt.clock, tt.idGenerator)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceCreateShipment(t *testing.T) {
	t.Parallel()

	repository := newStubRepository()
	clock := stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)}
	service := newTestService(t, repository, clock, &stubSequenceIDGenerator{ids: []string{"shipment-1", "shipment-2"}})

	shipment, err := service.CreateShipment(context.Background(), CreateShipmentCommand{
		ReferenceNumber:     "REF-001",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              DriverInput{ID: "driver-1", Name: "Jane Doe"},
		Unit:                UnitInput{ID: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: int64Ptr(100_00),
		DriverRevenueMinor:  int64Ptr(60_00),
	})
	if err != nil {
		t.Fatalf("CreateShipment() error = %v", err)
	}

	if shipment.ID != "shipment-1" {
		t.Fatalf("Shipment.ID = %q, want %q", shipment.ID, "shipment-1")
	}
	if shipment.ReferenceNumber != "REF-001" {
		t.Fatalf("Shipment.ReferenceNumber = %q, want %q", shipment.ReferenceNumber, "REF-001")
	}
	if shipment.CurrentStatus != string(domainshipment.StatusPending) {
		t.Fatalf("Shipment.CurrentStatus = %q, want %q", shipment.CurrentStatus, domainshipment.StatusPending)
	}
	if !shipment.CreatedAt.Equal(clock.now) {
		t.Fatalf("Shipment.CreatedAt = %v, want %v", shipment.CreatedAt, clock.now)
	}
}

func TestServiceCreateShipmentPropagatesDuplicateReference(t *testing.T) {
	t.Parallel()

	repository := newStubRepository()
	clock := stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)}
	service := newTestService(t, repository, clock, &stubSequenceIDGenerator{ids: []string{"shipment-1", "shipment-2"}})

	command := CreateShipmentCommand{
		ReferenceNumber:     "REF-001",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              DriverInput{ID: "driver-1", Name: "Jane Doe"},
		Unit:                UnitInput{ID: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: int64Ptr(100_00),
		DriverRevenueMinor:  int64Ptr(60_00),
	}

	if _, err := service.CreateShipment(context.Background(), command); err != nil {
		t.Fatalf("CreateShipment() first call error = %v", err)
	}

	if _, err := service.CreateShipment(context.Background(), command); !errors.Is(err, ports.ErrDuplicateReference) {
		t.Fatalf("CreateShipment() second call error = %v, want %v", err, ports.ErrDuplicateReference)
	}
}

func TestServiceCreateShipmentRejectsMissingMoney(t *testing.T) {
	t.Parallel()

	service := newTestService(
		t,
		newStubRepository(),
		stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)},
		stubIDGenerator{id: "shipment-1"},
	)

	_, err := service.CreateShipment(context.Background(), CreateShipmentCommand{
		ReferenceNumber:     "REF-001",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              DriverInput{ID: "driver-1", Name: "Jane Doe"},
		Unit:                UnitInput{ID: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: nil,
		DriverRevenueMinor:  int64Ptr(60_00),
	})
	if !errors.Is(err, domainshipment.ErrInvalidMoney) {
		t.Fatalf("CreateShipment() error = %v, want %v", err, domainshipment.ErrInvalidMoney)
	}
}

func TestServiceCreateShipmentReturnsContextErrorBeforeGeneratingID(t *testing.T) {
	t.Parallel()

	idGenerator := &countingIDGenerator{id: "shipment-1"}
	service := newTestService(
		t,
		newStubRepository(),
		stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)},
		idGenerator,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.CreateShipment(ctx, CreateShipmentCommand{
		ReferenceNumber:     "REF-001",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              DriverInput{ID: "driver-1", Name: "Jane Doe"},
		Unit:                UnitInput{ID: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: int64Ptr(100_00),
		DriverRevenueMinor:  int64Ptr(60_00),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateShipment() error = %v, want %v", err, context.Canceled)
	}
	if idGenerator.calls != 0 {
		t.Fatalf("ID generator calls = %d, want 0", idGenerator.calls)
	}
}

func TestServiceGetShipment(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	repository := newStubRepository()
	seedShipment(t, repository, createdAt)

	service := newTestService(t, repository, stubClock{now: createdAt}, stubIDGenerator{id: "ignored"})

	shipment, err := service.GetShipment(context.Background(), GetShipmentQuery{
		ReferenceNumber: " REF-001 ",
	})
	if err != nil {
		t.Fatalf("GetShipment() error = %v", err)
	}

	if shipment.ID != "shipment-1" {
		t.Fatalf("Shipment.ID = %q, want %q", shipment.ID, "shipment-1")
	}
	if shipment.ReferenceNumber != "REF-001" {
		t.Fatalf("Shipment.ReferenceNumber = %q, want %q", shipment.ReferenceNumber, "REF-001")
	}
}

func TestServiceGetShipmentRejectsBlankReference(t *testing.T) {
	t.Parallel()

	service := newTestService(
		t,
		newStubRepository(),
		stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)},
		stubIDGenerator{id: "ignored"},
	)

	_, err := service.GetShipment(context.Background(), GetShipmentQuery{ReferenceNumber: "   "})
	if !errors.Is(err, domainshipment.ErrInvalidReferenceNumber) {
		t.Fatalf("GetShipment() error = %v, want %v", err, domainshipment.ErrInvalidReferenceNumber)
	}
}

func TestServiceGetShipmentReturnsNotFound(t *testing.T) {
	t.Parallel()

	service := newTestService(
		t,
		newStubRepository(),
		stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)},
		stubIDGenerator{id: "ignored"},
	)

	_, err := service.GetShipment(context.Background(), GetShipmentQuery{ReferenceNumber: "REF-404"})
	if !errors.Is(err, ports.ErrShipmentNotFound) {
		t.Fatalf("GetShipment() error = %v, want %v", err, ports.ErrShipmentNotFound)
	}
}

func TestServiceAddStatusEventUsesClock(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	repository := newStubRepository()
	seedShipment(t, repository, createdAt)

	clock := stubClock{now: createdAt.Add(2 * time.Hour)}
	service := newTestService(t, repository, clock, stubIDGenerator{id: "ignored"})

	result, err := service.AddStatusEvent(context.Background(), AddShipmentStatusEventCommand{
		ReferenceNumber: "REF-001",
		Status:          "picked_up",
	})
	if err != nil {
		t.Fatalf("AddStatusEvent() error = %v", err)
	}

	if result.Event.Status != string(domainshipment.StatusPickedUp) {
		t.Fatalf("Event.Status = %q, want %q", result.Event.Status, domainshipment.StatusPickedUp)
	}
	if !result.Event.OccurredAt.Equal(clock.now) {
		t.Fatalf("Event.OccurredAt = %v, want %v", result.Event.OccurredAt, clock.now)
	}
	if result.Shipment.CurrentStatus != string(domainshipment.StatusPickedUp) {
		t.Fatalf("Shipment.CurrentStatus = %q, want %q", result.Shipment.CurrentStatus, domainshipment.StatusPickedUp)
	}
}

func TestServiceAddStatusEventRejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	repository := newStubRepository()
	seedShipment(t, repository, createdAt)

	service := newTestService(t, repository, stubClock{now: createdAt.Add(time.Hour)}, stubIDGenerator{id: "ignored"})

	_, err := service.AddStatusEvent(context.Background(), AddShipmentStatusEventCommand{
		ReferenceNumber: "REF-001",
		Status:          " picked_up ",
	})
	if !errors.Is(err, domainshipment.ErrInvalidStatus) {
		t.Fatalf("AddStatusEvent() error = %v, want %v", err, domainshipment.ErrInvalidStatus)
	}
}

func TestServiceAddStatusEventReturnsNotFound(t *testing.T) {
	t.Parallel()

	service := newTestService(
		t,
		newStubRepository(),
		stubClock{now: time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC)},
		stubIDGenerator{id: "ignored"},
	)

	_, err := service.AddStatusEvent(context.Background(), AddShipmentStatusEventCommand{
		ReferenceNumber: "REF-404",
		Status:          "picked_up",
	})
	if !errors.Is(err, ports.ErrShipmentNotFound) {
		t.Fatalf("AddStatusEvent() error = %v, want %v", err, ports.ErrShipmentNotFound)
	}
}

func TestServiceGetShipmentHistory(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	repository := newStubRepository()
	seedShipment(t, repository, createdAt)

	service := newTestService(t, repository, stubClock{now: createdAt.Add(2 * time.Hour)}, stubIDGenerator{id: "ignored"})
	if _, err := service.AddStatusEvent(context.Background(), AddShipmentStatusEventCommand{
		ReferenceNumber: "REF-001",
		Status:          "picked_up",
	}); err != nil {
		t.Fatalf("AddStatusEvent() error = %v", err)
	}

	history, err := service.GetShipmentHistory(context.Background(), GetShipmentHistoryQuery{
		ReferenceNumber: "REF-001",
	})
	if err != nil {
		t.Fatalf("GetShipmentHistory() error = %v", err)
	}

	if history.ShipmentID != "shipment-1" {
		t.Fatalf("History.ShipmentID = %q, want %q", history.ShipmentID, "shipment-1")
	}
	if history.ReferenceNumber != "REF-001" {
		t.Fatalf("History.ReferenceNumber = %q, want %q", history.ReferenceNumber, "REF-001")
	}
	if len(history.Events) != 2 {
		t.Fatalf("len(History.Events) = %d, want 2", len(history.Events))
	}
	if history.Events[0].Status != string(domainshipment.StatusPending) {
		t.Fatalf("History.Events[0].Status = %q, want %q", history.Events[0].Status, domainshipment.StatusPending)
	}
	if history.Events[1].Status != string(domainshipment.StatusPickedUp) {
		t.Fatalf("History.Events[1].Status = %q, want %q", history.Events[1].Status, domainshipment.StatusPickedUp)
	}
}

func TestServiceGetShipmentHistoryReturnsNotFound(t *testing.T) {
	t.Parallel()

	service := newTestService(
		t,
		newStubRepository(),
		stubClock{now: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)},
		stubIDGenerator{id: "ignored"},
	)

	_, err := service.GetShipmentHistory(context.Background(), GetShipmentHistoryQuery{
		ReferenceNumber: "REF-404",
	})
	if !errors.Is(err, ports.ErrShipmentNotFound) {
		t.Fatalf("GetShipmentHistory() error = %v, want %v", err, ports.ErrShipmentNotFound)
	}
}

type stubRepository struct {
	shipmentsByID        map[string]*domainshipment.Shipment
	shipmentsByReference map[string]string
}

func newStubRepository() *stubRepository {
	return &stubRepository{
		shipmentsByID:        make(map[string]*domainshipment.Shipment),
		shipmentsByReference: make(map[string]string),
	}
}

func (r *stubRepository) Create(_ context.Context, shipment *domainshipment.Shipment) error {
	if _, exists := r.shipmentsByID[shipment.ID()]; exists {
		return ports.ErrDuplicateShipmentID
	}
	if _, exists := r.shipmentsByReference[shipment.ReferenceNumber()]; exists {
		return ports.ErrDuplicateReference
	}

	r.shipmentsByID[shipment.ID()] = shipment
	r.shipmentsByReference[shipment.ReferenceNumber()] = shipment.ID()
	return nil
}

func (r *stubRepository) GetByReference(_ context.Context, referenceNumber string) (*domainshipment.Shipment, error) {
	id, ok := r.shipmentsByReference[referenceNumber]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	shipment, ok := r.shipmentsByID[id]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	return shipment, nil
}

func (r *stubRepository) UpdateByReference(_ context.Context, referenceNumber string, updateFn ports.ShipmentUpdateFn) (*domainshipment.Shipment, error) {
	id, ok := r.shipmentsByReference[referenceNumber]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	shipment, ok := r.shipmentsByID[id]
	if !ok {
		return nil, ports.ErrShipmentNotFound
	}

	if err := updateFn(shipment); err != nil {
		return nil, err
	}

	return shipment, nil
}

type stubClock struct {
	now time.Time
}

func (c stubClock) Now() time.Time {
	return c.now
}

type stubIDGenerator struct {
	id string
}

func (g stubIDGenerator) NewID() string {
	return g.id
}

type stubSequenceIDGenerator struct {
	ids   []string
	index int
}

func (g *stubSequenceIDGenerator) NewID() string {
	if len(g.ids) == 0 {
		return ""
	}
	if g.index >= len(g.ids) {
		return g.ids[len(g.ids)-1]
	}

	id := g.ids[g.index]
	g.index++
	return id
}

type countingIDGenerator struct {
	id    string
	calls int
}

func (g *countingIDGenerator) NewID() string {
	g.calls++
	return g.id
}

func newTestService(t *testing.T, repository ports.ShipmentRepository, clock ports.Clock, idGenerator ports.IDGenerator) *Service {
	t.Helper()

	service, err := New(repository, clock, idGenerator)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return service
}

func seedShipment(t *testing.T, repository *stubRepository, createdAt time.Time) {
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
		ID:              "shipment-1",
		ReferenceNumber: "REF-001",
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

	if err := repository.Create(context.Background(), shipment); err != nil {
		t.Fatalf("repository.Create() error = %v", err)
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
