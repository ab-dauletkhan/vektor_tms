package application

import "time"

type DriverInput struct {
	ID   string
	Name string
}

type UnitInput struct {
	ID                 string
	RegistrationNumber string
}

type CreateShipmentCommand struct {
	ReferenceNumber     string
	Origin              string
	Destination         string
	Driver              DriverInput
	Unit                UnitInput
	ShipmentAmountMinor *int64
	DriverRevenueMinor  *int64
}

type GetShipmentQuery struct {
	ReferenceNumber string
}

type AddShipmentStatusEventCommand struct {
	ReferenceNumber string
	Status          string
}

type GetShipmentHistoryQuery struct {
	ReferenceNumber string
}

type Driver struct {
	ID   string
	Name string
}

type Unit struct {
	ID                 string
	RegistrationNumber string
}

type Shipment struct {
	ID                  string
	ReferenceNumber     string
	Origin              string
	Destination         string
	CurrentStatus       string
	Driver              Driver
	Unit                Unit
	ShipmentAmountMinor int64
	DriverRevenueMinor  int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ShipmentEvent struct {
	Sequence   uint32
	Status     string
	OccurredAt time.Time
}

type ShipmentHistory struct {
	ShipmentID      string
	ReferenceNumber string
	Events          []ShipmentEvent
}

type AddStatusEventResult struct {
	Shipment Shipment
	Event    ShipmentEvent
}
