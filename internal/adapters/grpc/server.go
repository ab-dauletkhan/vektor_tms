package grpcadapter

import (
	"context"

	shipmentv1 "github.com/ab-dauletkhan/vektor_tms/api/proto/shipment/v1"
	"github.com/ab-dauletkhan/vektor_tms/internal/application"
)

type Server struct {
	shipmentv1.UnimplementedShipmentServiceServer

	application *application.Service
}

func NewServer(applicationService *application.Service) (*Server, error) {
	if applicationService == nil {
		return nil, errNilApplicationService
	}

	return &Server{
		application: applicationService,
	}, nil
}

func (s *Server) CreateShipment(ctx context.Context, req *shipmentv1.CreateShipmentRequest) (*shipmentv1.CreateShipmentResponse, error) {
	command, err := toCreateShipmentCommand(req)
	if err != nil {
		return nil, toGRPCError(err)
	}

	shipment, err := s.application.CreateShipment(ctx, command)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if shipment == nil {
		return nil, toGRPCError(errNilShipmentResponse)
	}

	mappedShipment, err := toProtoShipment(*shipment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &shipmentv1.CreateShipmentResponse{Shipment: mappedShipment}, nil
}

func (s *Server) GetShipment(ctx context.Context, req *shipmentv1.GetShipmentRequest) (*shipmentv1.GetShipmentResponse, error) {
	query, err := toGetShipmentQuery(req)
	if err != nil {
		return nil, toGRPCError(err)
	}

	shipment, err := s.application.GetShipment(ctx, query)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if shipment == nil {
		return nil, toGRPCError(errNilShipmentResponse)
	}

	mappedShipment, err := toProtoShipment(*shipment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &shipmentv1.GetShipmentResponse{Shipment: mappedShipment}, nil
}

func (s *Server) AddShipmentStatusEvent(ctx context.Context, req *shipmentv1.AddShipmentStatusEventRequest) (*shipmentv1.AddShipmentStatusEventResponse, error) {
	command, err := toAddStatusEventCommand(req)
	if err != nil {
		return nil, toGRPCError(err)
	}

	result, err := s.application.AddStatusEvent(ctx, command)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if result == nil {
		return nil, toGRPCError(errNilShipmentResponse)
	}

	mappedShipment, err := toProtoShipment(result.Shipment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	mappedEvent, err := toProtoEvent(result.Event)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &shipmentv1.AddShipmentStatusEventResponse{
		Shipment: mappedShipment,
		Event:    mappedEvent,
	}, nil
}

func (s *Server) GetShipmentHistory(ctx context.Context, req *shipmentv1.GetShipmentHistoryRequest) (*shipmentv1.GetShipmentHistoryResponse, error) {
	query, err := toGetShipmentHistoryQuery(req)
	if err != nil {
		return nil, toGRPCError(err)
	}

	history, err := s.application.GetShipmentHistory(ctx, query)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if history == nil {
		return nil, toGRPCError(errNilHistoryResponse)
	}

	response, err := toProtoHistory(*history)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return response, nil
}
