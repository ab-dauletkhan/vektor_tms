package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	shipmentv1 "github.com/ab-dauletkhan/vektor_tms/api/proto/shipment/v1"
	clockadapter "github.com/ab-dauletkhan/vektor_tms/internal/adapters/clock"
	grpcadapter "github.com/ab-dauletkhan/vektor_tms/internal/adapters/grpc"
	idadapter "github.com/ab-dauletkhan/vektor_tms/internal/adapters/id"
	"github.com/ab-dauletkhan/vektor_tms/internal/adapters/repository/memory"
	"github.com/ab-dauletkhan/vektor_tms/internal/application"
)

const gracefulShutdownTimeout = 30 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	address := getEnv("SHIPMENTD_ADDR", ":8080")

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("listen %q: %v", address, err)
	}

	applicationService, err := application.New(
		memory.New(),
		clockadapter.NewSystem(),
		idadapter.NewGenerator(),
	)
	if err != nil {
		log.Fatalf("create application service: %v", err)
	}

	shipmentServer, err := grpcadapter.NewServer(applicationService)
	if err != nil {
		log.Fatalf("create gRPC server adapter: %v", err)
	}

	logger := log.Default()
	grpcServer := grpc.NewServer(grpcadapter.ServerOptions(logger)...)
	shipmentv1.RegisterShipmentServiceServer(grpcServer, shipmentServer)

	go func() {
		<-ctx.Done()
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		timer := time.NewTimer(gracefulShutdownTimeout)
		defer timer.Stop()

		select {
		case <-done:
		case <-timer.C:
			logger.Printf("forcing gRPC shutdown after %s", gracefulShutdownTimeout)
			grpcServer.Stop()
			<-done
		}
	}()

	logger.Printf("shipmentd listening on %s", address)

	if err := grpcServer.Serve(listener); err != nil && ctx.Err() == nil {
		log.Fatalf("serve gRPC: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
