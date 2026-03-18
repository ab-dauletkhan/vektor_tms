package grpcadapter

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

var (
	errNilApplicationService = errors.New("nil application service")
	errNilRequest            = invalidArgumentError("request is required")
	errNilShipmentResponse   = errors.New("shipment response is required")
	errNilHistoryResponse    = errors.New("shipment history response is required")
	errUnknownStatus         = errors.New("unknown shipment status")
)

type invalidArgumentError string

func (e invalidArgumentError) Error() string {
	return string(e)
}

func (invalidArgumentError) IsInvalidArgument() bool {
	return true
}

type invalidArgumentClassifier interface {
	IsInvalidArgument() bool
}

type notFoundClassifier interface {
	IsNotFound() bool
}

type alreadyExistsClassifier interface {
	IsAlreadyExists() bool
}

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, context.Canceled):
		return grpcstatus.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return grpcstatus.Error(codes.DeadlineExceeded, err.Error())
	case isNotFoundError(err):
		return grpcstatus.Error(codes.NotFound, err.Error())
	case isAlreadyExistsError(err):
		return grpcstatus.Error(codes.AlreadyExists, err.Error())
	case isInvalidArgumentError(err):
		return grpcstatus.Error(codes.InvalidArgument, err.Error())
	default:
		return grpcstatus.Error(codes.Internal, "internal server error")
	}
}

func isInvalidArgumentError(err error) bool {
	var target invalidArgumentClassifier
	return errors.As(err, &target) && target.IsInvalidArgument()
}

func isNotFoundError(err error) bool {
	var target notFoundClassifier
	return errors.As(err, &target) && target.IsNotFound()
}

func isAlreadyExistsError(err error) bool {
	var target alreadyExistsClassifier
	return errors.As(err, &target) && target.IsAlreadyExists()
}
