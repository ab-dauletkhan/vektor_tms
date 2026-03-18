package shipment

import (
	"errors"
	"testing"
	"time"
)

func TestNewShipmentCreatesPendingHistory(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 12, 30, 0, 0, time.FixedZone("ALMT", 5*60*60))
	sh := newTestShipment(t, createdAt)

	if got := sh.ID(); got != "shipment-1" {
		t.Fatalf("ID() = %q, want %q", got, "shipment-1")
	}
	if got := sh.ReferenceNumber(); got != "REF-001" {
		t.Fatalf("ReferenceNumber() = %q, want %q", got, "REF-001")
	}
	if got := sh.Origin(); got != "Almaty" {
		t.Fatalf("Origin() = %q, want %q", got, "Almaty")
	}
	if got := sh.Destination(); got != "Astana" {
		t.Fatalf("Destination() = %q, want %q", got, "Astana")
	}
	if got := sh.Driver().ID(); got != "driver-1" {
		t.Fatalf("Driver().ID() = %q, want %q", got, "driver-1")
	}
	if got := sh.Driver().Name(); got != "Jane Doe" {
		t.Fatalf("Driver().Name() = %q, want %q", got, "Jane Doe")
	}
	if got := sh.Unit().ID(); got != "unit-1" {
		t.Fatalf("Unit().ID() = %q, want %q", got, "unit-1")
	}
	if got := sh.Unit().RegistrationNumber(); got != "123ABC02" {
		t.Fatalf("Unit().RegistrationNumber() = %q, want %q", got, "123ABC02")
	}
	if got := sh.CurrentStatus(); got != StatusPending {
		t.Fatalf("CurrentStatus() = %q, want %q", got, StatusPending)
	}

	wantCreatedAt := createdAt.UTC()
	if got := sh.CreatedAt(); !got.Equal(wantCreatedAt) {
		t.Fatalf("CreatedAt() = %v, want %v", got, wantCreatedAt)
	}
	if got := sh.UpdatedAt(); !got.Equal(wantCreatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", got, wantCreatedAt)
	}

	events := sh.Events()
	if len(events) != 1 {
		t.Fatalf("len(Events()) = %d, want 1", len(events))
	}
	if got := events[0].Sequence(); got != 1 {
		t.Fatalf("Events()[0].Sequence() = %d, want 1", got)
	}
	if got := events[0].Status(); got != StatusPending {
		t.Fatalf("Events()[0].Status() = %q, want %q", got, StatusPending)
	}
	if got := events[0].OccurredAt(); !got.Equal(wantCreatedAt) {
		t.Fatalf("Events()[0].OccurredAt() = %v, want %v", got, wantCreatedAt)
	}
}

func TestNewShipmentRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	validDriver := mustDriver(t, "driver-1", "Jane Doe")
	validUnit := mustUnit(t, "unit-1", "123ABC02")
	validAmount := mustMoney(t, 100_00)
	validRevenue := mustMoney(t, 60_00)
	var zeroValueMoney Money

	testCases := []struct {
		name    string
		params  NewParams
		wantErr error
	}{
		{
			name: "blank shipment id",
			params: NewParams{
				ID:              "   ",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidShipmentID,
		},
		{
			name: "blank reference number",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "   ",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidReferenceNumber,
		},
		{
			name: "same origin and destination after trimming",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          " Almaty ",
				Destination:     "Almaty",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrOriginEqualsDestination,
		},
		{
			name: "invalid driver",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          Driver{},
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidDriver,
		},
		{
			name: "invalid unit",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            Unit{},
				ShipmentAmount:  validAmount,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidUnit,
		},
		{
			name: "zero created_at",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				DriverRevenue:   validRevenue,
			},
			wantErr: ErrInvalidCreatedAt,
		},
		{
			name: "driver revenue exceeds shipment amount",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				DriverRevenue:   mustMoney(t, 101_00),
				CreatedAt:       createdAt,
			},
			wantErr: ErrRevenueExceedsAmount,
		},
		{
			name: "missing shipment amount",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidMoney,
		},
		{
			name: "missing driver revenue",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  validAmount,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidMoney,
		},
		{
			name: "zero value shipment amount is rejected",
			params: NewParams{
				ID:              "shipment-1",
				ReferenceNumber: "REF-001",
				Origin:          "Almaty",
				Destination:     "Astana",
				Driver:          validDriver,
				Unit:            validUnit,
				ShipmentAmount:  &zeroValueMoney,
				DriverRevenue:   validRevenue,
				CreatedAt:       createdAt,
			},
			wantErr: ErrInvalidMoney,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewShipment(tc.params)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("NewShipment() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestShipmentAddStatusEventValidTransitions(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	sh := newTestShipment(t, createdAt)

	transitions := []struct {
		next       Status
		occurredAt time.Time
	}{
		{next: StatusPickedUp, occurredAt: createdAt.Add(30 * time.Minute)},
		{next: StatusInTransit, occurredAt: createdAt.Add(2 * time.Hour)},
		{next: StatusDelivered, occurredAt: createdAt.Add(6 * time.Hour)},
	}

	for i, transition := range transitions {
		event, err := sh.AddStatusEvent(transition.next, transition.occurredAt)
		if err != nil {
			t.Fatalf("AddStatusEvent(%q) error = %v", transition.next, err)
		}

		wantSequence := uint32(i + 2)
		if got := event.Sequence(); got != wantSequence {
			t.Fatalf("event.Sequence() = %d, want %d", got, wantSequence)
		}
		if got := event.Status(); got != transition.next {
			t.Fatalf("event.Status() = %q, want %q", got, transition.next)
		}
		if got := event.OccurredAt(); !got.Equal(transition.occurredAt.UTC()) {
			t.Fatalf("event.OccurredAt() = %v, want %v", got, transition.occurredAt.UTC())
		}
		if got := sh.CurrentStatus(); got != transition.next {
			t.Fatalf("CurrentStatus() = %q, want %q", got, transition.next)
		}
		if got := sh.UpdatedAt(); !got.Equal(transition.occurredAt.UTC()) {
			t.Fatalf("UpdatedAt() = %v, want %v", got, transition.occurredAt.UTC())
		}
	}

	if got := len(sh.Events()); got != 4 {
		t.Fatalf("len(Events()) = %d, want 4", got)
	}
}

func TestShipmentAddStatusEventAllowsPendingCancellation(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	sh := newTestShipment(t, createdAt)

	event, err := sh.AddStatusEvent(StatusCancelled, createdAt.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("AddStatusEvent(%q) error = %v", StatusCancelled, err)
	}

	if got := event.Sequence(); got != 2 {
		t.Fatalf("event.Sequence() = %d, want 2", got)
	}
	if got := sh.CurrentStatus(); got != StatusCancelled {
		t.Fatalf("CurrentStatus() = %q, want %q", got, StatusCancelled)
	}
	if got := sh.UpdatedAt(); !got.Equal(createdAt.Add(30 * time.Minute)) {
		t.Fatalf("UpdatedAt() = %v, want %v", got, createdAt.Add(30*time.Minute))
	}
}

func TestShipmentAddStatusEventRejectsInvalidTransitions(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		setup    func(t *testing.T) *Shipment
		next     Status
		at       time.Time
		wantErr  error
		wantSize int
	}{
		{
			name: "duplicate current status",
			setup: func(t *testing.T) *Shipment {
				return newTestShipment(t, createdAt)
			},
			next:     StatusPending,
			at:       createdAt.Add(30 * time.Minute),
			wantErr:  ErrDuplicateStatus,
			wantSize: 1,
		},
		{
			name: "skips lifecycle steps",
			setup: func(t *testing.T) *Shipment {
				return newTestShipment(t, createdAt)
			},
			next:     StatusDelivered,
			at:       createdAt.Add(30 * time.Minute),
			wantErr:  ErrInvalidTransition,
			wantSize: 1,
		},
		{
			name: "cannot cancel after pickup",
			setup: func(t *testing.T) *Shipment {
				sh := newTestShipment(t, createdAt)
				if _, err := sh.AddStatusEvent(StatusPickedUp, createdAt.Add(30*time.Minute)); err != nil {
					t.Fatalf("setup AddStatusEvent() error = %v", err)
				}
				return sh
			},
			next:     StatusCancelled,
			at:       createdAt.Add(45 * time.Minute),
			wantErr:  ErrInvalidTransition,
			wantSize: 2,
		},
		{
			name: "out of order timestamp",
			setup: func(t *testing.T) *Shipment {
				sh := newTestShipment(t, createdAt)
				if _, err := sh.AddStatusEvent(StatusPickedUp, createdAt.Add(2*time.Hour)); err != nil {
					t.Fatalf("setup AddStatusEvent() error = %v", err)
				}
				return sh
			},
			next:     StatusInTransit,
			at:       createdAt.Add(90 * time.Minute),
			wantErr:  ErrEventOutOfOrder,
			wantSize: 2,
		},
		{
			name: "invalid status",
			setup: func(t *testing.T) *Shipment {
				return newTestShipment(t, createdAt)
			},
			next:     Status("unknown"),
			at:       createdAt.Add(30 * time.Minute),
			wantErr:  ErrInvalidStatus,
			wantSize: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sh := tc.setup(t)
			beforeStatus := sh.CurrentStatus()
			beforeUpdatedAt := sh.UpdatedAt()

			_, err := sh.AddStatusEvent(tc.next, tc.at)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("AddStatusEvent() error = %v, want %v", err, tc.wantErr)
			}
			if got := len(sh.Events()); got != tc.wantSize {
				t.Fatalf("len(Events()) = %d, want %d", got, tc.wantSize)
			}
			if got := sh.CurrentStatus(); got != beforeStatus {
				t.Fatalf("CurrentStatus() = %q after invalid transition, want %q", got, beforeStatus)
			}
			if got := sh.UpdatedAt(); !got.Equal(beforeUpdatedAt) {
				t.Fatalf("UpdatedAt() = %v after invalid transition, want %v", got, beforeUpdatedAt)
			}
		})
	}
}

func TestShipmentEventsReturnsCopy(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	sh := newTestShipment(t, createdAt)

	events := sh.Events()
	events[0] = mustEvent(t, 1, StatusCancelled, createdAt.Add(time.Hour))

	if got := sh.CurrentStatus(); got != StatusPending {
		t.Fatalf("CurrentStatus() after external history mutation = %q, want %q", got, StatusPending)
	}
	if got := sh.Events()[0].Status(); got != StatusPending {
		t.Fatalf("Events()[0].Status() after external history mutation = %q, want %q", got, StatusPending)
	}
}

func TestShipmentTerminalStatesRejectFurtherTransitions(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		setup         func(t *testing.T) *Shipment
		next          Status
		expectedState Status
	}{
		{
			name: "delivered is terminal",
			setup: func(t *testing.T) *Shipment {
				sh := newTestShipment(t, createdAt)
				for _, status := range []Status{StatusPickedUp, StatusInTransit, StatusDelivered} {
					if _, err := sh.AddStatusEvent(status, createdAt.Add(time.Duration(len(sh.Events()))*time.Hour)); err != nil {
						t.Fatalf("setup AddStatusEvent(%q) error = %v", status, err)
					}
				}
				return sh
			},
			next:          StatusCancelled,
			expectedState: StatusDelivered,
		},
		{
			name: "cancelled is terminal",
			setup: func(t *testing.T) *Shipment {
				sh := newTestShipment(t, createdAt)
				if _, err := sh.AddStatusEvent(StatusCancelled, createdAt.Add(time.Hour)); err != nil {
					t.Fatalf("setup AddStatusEvent(%q) error = %v", StatusCancelled, err)
				}
				return sh
			},
			next:          StatusPickedUp,
			expectedState: StatusCancelled,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sh := tc.setup(t)
			beforeUpdatedAt := sh.UpdatedAt()

			_, err := sh.AddStatusEvent(tc.next, beforeUpdatedAt.Add(time.Hour))
			if !errors.Is(err, ErrInvalidTransition) {
				t.Fatalf("AddStatusEvent() error = %v, want %v", err, ErrInvalidTransition)
			}
			if got := sh.CurrentStatus(); got != tc.expectedState {
				t.Fatalf("CurrentStatus() = %q, want %q", got, tc.expectedState)
			}
			if got := sh.UpdatedAt(); !got.Equal(beforeUpdatedAt) {
				t.Fatalf("UpdatedAt() = %v, want %v", got, beforeUpdatedAt)
			}
		})
	}
}

func TestRehydrateRebuildsShipmentFromHistory(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	history := []Event{
		mustEvent(t, 1, StatusPending, createdAt),
		mustEvent(t, 2, StatusPickedUp, createdAt.Add(30*time.Minute)),
		mustEvent(t, 3, StatusInTransit, createdAt.Add(2*time.Hour)),
	}

	sh, err := Rehydrate(RehydrateParams{
		ID:              "shipment-1",
		ReferenceNumber: "REF-001",
		Origin:          " Almaty ",
		Destination:     " Astana ",
		Driver:          mustDriver(t, "driver-1", "Jane Doe"),
		Unit:            mustUnit(t, "unit-1", "123ABC02"),
		ShipmentAmount:  mustMoney(t, 100_00),
		DriverRevenue:   mustMoney(t, 60_00),
		Events:          history,
	})
	if err != nil {
		t.Fatalf("Rehydrate() error = %v", err)
	}

	if got := sh.CurrentStatus(); got != StatusInTransit {
		t.Fatalf("CurrentStatus() = %q, want %q", got, StatusInTransit)
	}
	if got := sh.CreatedAt(); !got.Equal(createdAt) {
		t.Fatalf("CreatedAt() = %v, want %v", got, createdAt)
	}
	if got := sh.UpdatedAt(); !got.Equal(createdAt.Add(2 * time.Hour)) {
		t.Fatalf("UpdatedAt() = %v, want %v", got, createdAt.Add(2*time.Hour))
	}

	history[2] = mustEvent(t, 3, StatusDelivered, createdAt.Add(3*time.Hour))
	if got := sh.CurrentStatus(); got != StatusInTransit {
		t.Fatalf("CurrentStatus() after external slice mutation = %q, want %q", got, StatusInTransit)
	}
}

func TestRehydrateRejectsInvalidHistory(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	base := RehydrateParams{
		ID:              "shipment-1",
		ReferenceNumber: "REF-001",
		Origin:          "Almaty",
		Destination:     "Astana",
		Driver:          mustDriver(t, "driver-1", "Jane Doe"),
		Unit:            mustUnit(t, "unit-1", "123ABC02"),
		ShipmentAmount:  mustMoney(t, 100_00),
		DriverRevenue:   mustMoney(t, 60_00),
	}
	var zeroValueMoney Money

	testCases := []struct {
		name    string
		params  func() RehydrateParams
		history []Event
		wantErr error
	}{
		{
			name:    "empty history",
			params:  func() RehydrateParams { return base },
			history: nil,
			wantErr: ErrEmptyEventHistory,
		},
		{
			name:   "initial status must be pending",
			params: func() RehydrateParams { return base },
			history: []Event{
				mustEvent(t, 1, StatusPickedUp, createdAt),
			},
			wantErr: ErrInvalidInitialStatus,
		},
		{
			name:   "sequence must be contiguous",
			params: func() RehydrateParams { return base },
			history: []Event{
				mustEvent(t, 1, StatusPending, createdAt),
				mustEvent(t, 3, StatusPickedUp, createdAt.Add(time.Hour)),
			},
			wantErr: ErrInvalidEventSequence,
		},
		{
			name:   "history must be chronological",
			params: func() RehydrateParams { return base },
			history: []Event{
				mustEvent(t, 1, StatusPending, createdAt),
				mustEvent(t, 2, StatusPickedUp, createdAt.Add(2*time.Hour)),
				mustEvent(t, 3, StatusInTransit, createdAt.Add(time.Hour)),
			},
			wantErr: ErrEventOutOfOrder,
		},
		{
			name:   "history must respect transitions",
			params: func() RehydrateParams { return base },
			history: []Event{
				mustEvent(t, 1, StatusPending, createdAt),
				mustEvent(t, 2, StatusDelivered, createdAt.Add(time.Hour)),
			},
			wantErr: ErrInvalidTransition,
		},
		{
			name: "missing shipment amount",
			params: func() RehydrateParams {
				params := base
				params.ShipmentAmount = nil
				return params
			},
			history: []Event{
				mustEvent(t, 1, StatusPending, createdAt),
			},
			wantErr: ErrInvalidMoney,
		},
		{
			name: "zero value driver revenue is rejected",
			params: func() RehydrateParams {
				params := base
				params.DriverRevenue = &zeroValueMoney
				return params
			},
			history: []Event{
				mustEvent(t, 1, StatusPending, createdAt),
			},
			wantErr: ErrInvalidMoney,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			params := tc.params()
			params.Events = tc.history

			_, err := Rehydrate(params)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Rehydrate() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestRehydrateNormalizesEventTimesToUTC(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("ALMT", 5*60*60)
	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, loc)

	sh, err := Rehydrate(RehydrateParams{
		ID:              "shipment-1",
		ReferenceNumber: "REF-001",
		Origin:          "Almaty",
		Destination:     "Astana",
		Driver:          mustDriver(t, "driver-1", "Jane Doe"),
		Unit:            mustUnit(t, "unit-1", "123ABC02"),
		ShipmentAmount:  mustMoney(t, 100_00),
		DriverRevenue:   mustMoney(t, 60_00),
		Events: []Event{
			mustEvent(t, 1, StatusPending, createdAt),
		},
	})
	if err != nil {
		t.Fatalf("Rehydrate() error = %v", err)
	}

	if got := sh.CreatedAt(); got.Location() != time.UTC {
		t.Fatalf("CreatedAt() location = %v, want UTC", got.Location())
	}
}

func newTestShipment(t *testing.T, createdAt time.Time) *Shipment {
	t.Helper()

	sh, err := NewShipment(NewParams{
		ID:              " shipment-1 ",
		ReferenceNumber: " REF-001 ",
		Origin:          " Almaty ",
		Destination:     " Astana ",
		Driver:          mustDriver(t, " driver-1 ", " Jane Doe "),
		Unit:            mustUnit(t, " unit-1 ", " 123ABC02 "),
		ShipmentAmount:  mustMoney(t, 100_00),
		DriverRevenue:   mustMoney(t, 60_00),
		CreatedAt:       createdAt,
	})
	if err != nil {
		t.Fatalf("NewShipment() error = %v", err)
	}

	return sh
}

func mustDriver(t *testing.T, id, name string) Driver {
	t.Helper()

	driver, err := NewDriver(id, name)
	if err != nil {
		t.Fatalf("NewDriver() error = %v", err)
	}

	return driver
}

func mustUnit(t *testing.T, id, registrationNumber string) Unit {
	t.Helper()

	unit, err := NewUnit(id, registrationNumber)
	if err != nil {
		t.Fatalf("NewUnit() error = %v", err)
	}

	return unit
}

func mustMoney(t *testing.T, minorUnits int64) *Money {
	t.Helper()

	money, err := NewMoney(minorUnits)
	if err != nil {
		t.Fatalf("NewMoney() error = %v", err)
	}

	return &money
}

func mustEvent(t *testing.T, sequence uint32, status Status, occurredAt time.Time) Event {
	t.Helper()

	event, err := RehydrateEvent(sequence, status, occurredAt)
	if err != nil {
		t.Fatalf("RehydrateEvent() error = %v", err)
	}

	return event
}
