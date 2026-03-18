package grpcadapter

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	shipmentv1 "github.com/ab-dauletkhan/vektor_tms/api/proto/shipment/v1"
	"github.com/ab-dauletkhan/vektor_tms/internal/application"
	domainshipment "github.com/ab-dauletkhan/vektor_tms/internal/domain/shipment"
)

var (
	protoStatusByDomainStatus = map[domainshipment.Status]shipmentv1.ShipmentStatus{
		domainshipment.StatusPending:   shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PENDING,
		domainshipment.StatusPickedUp:  shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP,
		domainshipment.StatusInTransit: shipmentv1.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT,
		domainshipment.StatusDelivered: shipmentv1.ShipmentStatus_SHIPMENT_STATUS_DELIVERED,
		domainshipment.StatusCancelled: shipmentv1.ShipmentStatus_SHIPMENT_STATUS_CANCELLED,
	}
	domainStatusByProtoStatus = map[shipmentv1.ShipmentStatus]domainshipment.Status{
		shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PENDING:    domainshipment.StatusPending,
		shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP:  domainshipment.StatusPickedUp,
		shipmentv1.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT: domainshipment.StatusInTransit,
		shipmentv1.ShipmentStatus_SHIPMENT_STATUS_DELIVERED:  domainshipment.StatusDelivered,
		shipmentv1.ShipmentStatus_SHIPMENT_STATUS_CANCELLED:  domainshipment.StatusCancelled,
	}
)

func init() {
	validateStatusMappings()
}

func toCreateShipmentCommand(req *shipmentv1.CreateShipmentRequest) (application.CreateShipmentCommand, error) {
	if req == nil {
		return application.CreateShipmentCommand{}, errNilRequest
	}

	driver, err := toDriverInput(req.GetDriver())
	if err != nil {
		return application.CreateShipmentCommand{}, err
	}

	unit, err := toUnitInput(req.GetUnit())
	if err != nil {
		return application.CreateShipmentCommand{}, err
	}

	return application.CreateShipmentCommand{
		ReferenceNumber:     req.GetReferenceNumber(),
		Origin:              req.GetOrigin(),
		Destination:         req.GetDestination(),
		Driver:              driver,
		Unit:                unit,
		ShipmentAmountMinor: copyOptionalInt64(req.ShipmentAmountMinor),
		DriverRevenueMinor:  copyOptionalInt64(req.DriverRevenueMinor),
	}, nil
}

func toGetShipmentQuery(req *shipmentv1.GetShipmentRequest) (application.GetShipmentQuery, error) {
	if req == nil {
		return application.GetShipmentQuery{}, errNilRequest
	}

	return application.GetShipmentQuery{ReferenceNumber: req.GetReferenceNumber()}, nil
}

func toAddStatusEventCommand(req *shipmentv1.AddShipmentStatusEventRequest) (application.AddShipmentStatusEventCommand, error) {
	if req == nil {
		return application.AddShipmentStatusEventCommand{}, errNilRequest
	}

	status, err := fromProtoStatus(req.GetStatus())
	if err != nil {
		return application.AddShipmentStatusEventCommand{}, err
	}

	return application.AddShipmentStatusEventCommand{
		ReferenceNumber: req.GetReferenceNumber(),
		Status:          status,
	}, nil
}

func toGetShipmentHistoryQuery(req *shipmentv1.GetShipmentHistoryRequest) (application.GetShipmentHistoryQuery, error) {
	if req == nil {
		return application.GetShipmentHistoryQuery{}, errNilRequest
	}

	return application.GetShipmentHistoryQuery{ReferenceNumber: req.GetReferenceNumber()}, nil
}

func toDriverInput(driver *shipmentv1.Driver) (application.DriverInput, error) {
	if driver == nil {
		return application.DriverInput{}, domainshipment.ErrInvalidDriver
	}

	return application.DriverInput{
		ID:   driver.GetId(),
		Name: driver.GetName(),
	}, nil
}

func toUnitInput(unit *shipmentv1.Unit) (application.UnitInput, error) {
	if unit == nil {
		return application.UnitInput{}, domainshipment.ErrInvalidUnit
	}

	return application.UnitInput{
		ID:                 unit.GetId(),
		RegistrationNumber: unit.GetRegistrationNumber(),
	}, nil
}

func copyOptionalInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}

func fromProtoStatus(status shipmentv1.ShipmentStatus) (string, error) {
	if status == shipmentv1.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED {
		return "", domainshipment.ErrInvalidStatus
	}

	mapped, ok := domainStatusByProtoStatus[status]
	if !ok {
		return "", domainshipment.ErrInvalidStatus
	}

	return string(mapped), nil
}

func toProtoShipment(shipment application.Shipment) (*shipmentv1.Shipment, error) {
	status, err := toProtoStatus(shipment.CurrentStatus)
	if err != nil {
		return nil, err
	}

	return &shipmentv1.Shipment{
		Id:              shipment.ID,
		ReferenceNumber: shipment.ReferenceNumber,
		Origin:          shipment.Origin,
		Destination:     shipment.Destination,
		CurrentStatus:   status,
		Driver: &shipmentv1.Driver{
			Id:   shipment.Driver.ID,
			Name: shipment.Driver.Name,
		},
		Unit: &shipmentv1.Unit{
			Id:                 shipment.Unit.ID,
			RegistrationNumber: shipment.Unit.RegistrationNumber,
		},
		ShipmentAmountMinor: shipment.ShipmentAmountMinor,
		DriverRevenueMinor:  shipment.DriverRevenueMinor,
		CreatedAt:           timestamppb.New(shipment.CreatedAt),
		UpdatedAt:           timestamppb.New(shipment.UpdatedAt),
	}, nil
}

func toProtoEvent(event application.ShipmentEvent) (*shipmentv1.ShipmentEvent, error) {
	status, err := toProtoStatus(event.Status)
	if err != nil {
		return nil, err
	}

	return &shipmentv1.ShipmentEvent{
		Sequence:   event.Sequence,
		Status:     status,
		OccurredAt: timestamppb.New(event.OccurredAt),
	}, nil
}

func toProtoHistory(history application.ShipmentHistory) (*shipmentv1.GetShipmentHistoryResponse, error) {
	events := make([]*shipmentv1.ShipmentEvent, len(history.Events))
	for i, event := range history.Events {
		mapped, err := toProtoEvent(event)
		if err != nil {
			return nil, err
		}
		events[i] = mapped
	}

	return &shipmentv1.GetShipmentHistoryResponse{
		ShipmentId:      history.ShipmentID,
		ReferenceNumber: history.ReferenceNumber,
		Events:          events,
	}, nil
}

func toProtoStatus(status string) (shipmentv1.ShipmentStatus, error) {
	mapped, ok := protoStatusByDomainStatus[domainshipment.Status(status)]
	if !ok {
		return shipmentv1.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED, errUnknownStatus
	}

	return mapped, nil
}

func validateStatusMappings() {
	domainStatuses := domainshipment.AllStatuses()
	if len(protoStatusByDomainStatus) != len(domainStatuses) {
		panic(fmt.Sprintf("proto status mapping mismatch: got %d domain mappings, want %d", len(protoStatusByDomainStatus), len(domainStatuses)))
	}
	for _, status := range domainStatuses {
		if _, ok := protoStatusByDomainStatus[status]; !ok {
			panic(fmt.Sprintf("missing proto status mapping for domain status %q", status))
		}
	}

	protoStatusCount := 0
	for value := range shipmentv1.ShipmentStatus_name {
		status := shipmentv1.ShipmentStatus(value)
		if status == shipmentv1.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED {
			continue
		}

		protoStatusCount++
		if _, ok := domainStatusByProtoStatus[status]; !ok {
			panic(fmt.Sprintf("missing domain status mapping for proto status %s", status))
		}
	}
	if len(domainStatusByProtoStatus) != protoStatusCount {
		panic(fmt.Sprintf("domain status mapping mismatch: got %d proto mappings, want %d", len(domainStatusByProtoStatus), protoStatusCount))
	}
}
