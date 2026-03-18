package grpcadapter

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	shipmentv1 "github.com/ab-dauletkhan/vektor_tms/api/proto/shipment/v1"
	"github.com/ab-dauletkhan/vektor_tms/internal/adapters/repository/memory"
	"github.com/ab-dauletkhan/vektor_tms/internal/application"
)

func TestServerCreateGetAddEventAndHistory(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	pickedUpAt := createdAt.Add(2 * time.Hour)
	client, cleanup := newTestClient(
		t,
		&sequenceClock{times: []time.Time{createdAt, pickedUpAt}},
		&sequenceIDGenerator{ids: []string{"shipment-1"}},
	)
	defer cleanup()

	createResponse, err := client.CreateShipment(context.Background(), &shipmentv1.CreateShipmentRequest{
		ReferenceNumber:     "REF-001",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              &shipmentv1.Driver{Id: "driver-1", Name: "Jane Doe"},
		Unit:                &shipmentv1.Unit{Id: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: int64Ptr(100_00),
		DriverRevenueMinor:  int64Ptr(60_00),
	})
	if err != nil {
		t.Fatalf("CreateShipment() error = %v", err)
	}

	if got := createResponse.GetShipment().GetId(); got != "shipment-1" {
		t.Fatalf("CreateShipment().Shipment.Id = %q, want %q", got, "shipment-1")
	}
	if got := createResponse.GetShipment().GetCurrentStatus(); got != shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PENDING {
		t.Fatalf("CreateShipment().Shipment.CurrentStatus = %v, want %v", got, shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PENDING)
	}
	if got := createResponse.GetShipment().GetCreatedAt().AsTime(); !got.Equal(createdAt) {
		t.Fatalf("CreateShipment().Shipment.CreatedAt = %v, want %v", got, createdAt)
	}

	getResponse, err := client.GetShipment(context.Background(), &shipmentv1.GetShipmentRequest{
		ReferenceNumber: "REF-001",
	})
	if err != nil {
		t.Fatalf("GetShipment() error = %v", err)
	}
	if got := getResponse.GetShipment().GetReferenceNumber(); got != "REF-001" {
		t.Fatalf("GetShipment().Shipment.ReferenceNumber = %q, want %q", got, "REF-001")
	}

	addResponse, err := client.AddShipmentStatusEvent(context.Background(), &shipmentv1.AddShipmentStatusEventRequest{
		ReferenceNumber: "REF-001",
		Status:          shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP,
	})
	if err != nil {
		t.Fatalf("AddShipmentStatusEvent() error = %v", err)
	}
	if got := addResponse.GetEvent().GetStatus(); got != shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP {
		t.Fatalf("AddShipmentStatusEvent().Event.Status = %v, want %v", got, shipmentv1.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP)
	}
	if got := addResponse.GetEvent().GetOccurredAt().AsTime(); !got.Equal(pickedUpAt) {
		t.Fatalf("AddShipmentStatusEvent().Event.OccurredAt = %v, want %v", got, pickedUpAt)
	}

	historyResponse, err := client.GetShipmentHistory(context.Background(), &shipmentv1.GetShipmentHistoryRequest{
		ReferenceNumber: "REF-001",
	})
	if err != nil {
		t.Fatalf("GetShipmentHistory() error = %v", err)
	}
	if got := historyResponse.GetShipmentId(); got != "shipment-1" {
		t.Fatalf("GetShipmentHistory().ShipmentId = %q, want %q", got, "shipment-1")
	}
	if got := len(historyResponse.GetEvents()); got != 2 {
		t.Fatalf("len(GetShipmentHistory().Events) = %d, want 2", got)
	}
}

func TestServerMapsErrorsToGRPCCodes(t *testing.T) {
	t.Parallel()

	client, cleanup := newTestClient(
		t,
		&sequenceClock{times: []time.Time{time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)}},
		&sequenceIDGenerator{ids: []string{"shipment-1", "shipment-2"}},
	)
	defer cleanup()

	_, err := client.GetShipment(context.Background(), &shipmentv1.GetShipmentRequest{
		ReferenceNumber: "REF-404",
	})
	assertCode(t, err, codes.NotFound)

	_, err = client.CreateShipment(context.Background(), &shipmentv1.CreateShipmentRequest{
		ReferenceNumber:    "REF-001",
		Origin:             "Almaty",
		Destination:        "Astana",
		Driver:             &shipmentv1.Driver{Id: "driver-1", Name: "Jane Doe"},
		Unit:               &shipmentv1.Unit{Id: "unit-1", RegistrationNumber: "123ABC02"},
		DriverRevenueMinor: int64Ptr(60_00),
	})
	assertCode(t, err, codes.InvalidArgument)

	_, err = client.CreateShipment(context.Background(), &shipmentv1.CreateShipmentRequest{
		ReferenceNumber:     "REF-002",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              &shipmentv1.Driver{Id: "driver-1", Name: "Jane Doe"},
		Unit:                &shipmentv1.Unit{Id: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: int64Ptr(100_00),
		DriverRevenueMinor:  int64Ptr(60_00),
	})
	if err != nil {
		t.Fatalf("CreateShipment() setup error = %v", err)
	}

	_, err = client.CreateShipment(context.Background(), &shipmentv1.CreateShipmentRequest{
		ReferenceNumber:     "REF-002",
		Origin:              "Almaty",
		Destination:         "Astana",
		Driver:              &shipmentv1.Driver{Id: "driver-1", Name: "Jane Doe"},
		Unit:                &shipmentv1.Unit{Id: "unit-1", RegistrationNumber: "123ABC02"},
		ShipmentAmountMinor: int64Ptr(100_00),
		DriverRevenueMinor:  int64Ptr(60_00),
	})
	assertCode(t, err, codes.AlreadyExists)

	_, err = client.AddShipmentStatusEvent(context.Background(), &shipmentv1.AddShipmentStatusEventRequest{
		ReferenceNumber: "REF-002",
		Status:          shipmentv1.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED,
	})
	assertCode(t, err, codes.InvalidArgument)
}

func newTestClient(t *testing.T, clock applicationClock, idGenerator applicationIDGenerator) (shipmentv1.ShipmentServiceClient, func()) {
	t.Helper()

	appService, err := application.New(memory.New(), clock, idGenerator)
	if err != nil {
		t.Fatalf("application.New() error = %v", err)
	}

	server, err := NewServer(appService)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer(ServerOptions(log.New(io.Discard, "", 0))...)
	shipmentv1.RegisterShipmentServiceServer(grpcServer, server)

	go func() {
		if serveErr := grpcServer.Serve(listener); serveErr != nil && !errors.Is(serveErr, net.ErrClosed) {
			panic(serveErr)
		}
	}()

	connection, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient() error = %v", err)
	}

	cleanup := func() {
		_ = connection.Close()
		grpcServer.Stop()
		_ = listener.Close()
	}

	return shipmentv1.NewShipmentServiceClient(connection), cleanup
}

func assertCode(t *testing.T, err error, want codes.Code) {
	t.Helper()

	if err == nil {
		t.Fatalf("error = nil, want gRPC code %v", want)
	}

	if got := grpcstatus.Code(err); got != want {
		t.Fatalf("gRPC code = %v, want %v (error = %v)", got, want, err)
	}
}

type applicationClock interface {
	Now() time.Time
}

type applicationIDGenerator interface {
	NewID() string
}

type sequenceClock struct {
	times []time.Time
	index int
}

func (c *sequenceClock) Now() time.Time {
	if len(c.times) == 0 {
		return time.Time{}
	}
	if c.index >= len(c.times) {
		return c.times[len(c.times)-1]
	}

	current := c.times[c.index]
	c.index++
	return current
}

type sequenceIDGenerator struct {
	ids   []string
	index int
}

func (g *sequenceIDGenerator) NewID() string {
	if len(g.ids) == 0 {
		return ""
	}
	if g.index >= len(g.ids) {
		return g.ids[len(g.ids)-1]
	}

	current := g.ids[g.index]
	g.index++
	return current
}

func int64Ptr(value int64) *int64 {
	return &value
}
