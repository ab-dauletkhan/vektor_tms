package grpcadapter

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func ServerOptions(logger *log.Logger) []grpc.ServerOption {
	logger = normalizeLogger(logger)

	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor(logger),
			recoveryUnaryInterceptor(logger),
		),
	}
}

func loggingUnaryInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startedAt := time.Now()

		response, err := handler(ctx, req)
		logger.Printf(
			"grpc unary method=%s code=%s duration=%s",
			info.FullMethod,
			grpcstatus.Code(err),
			time.Since(startedAt),
		)

		return response, err
	}
}

func recoveryUnaryInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Printf(
					"grpc panic recovered method=%s panic=%v\n%s",
					info.FullMethod,
					recovered,
					debug.Stack(),
				)
				err = grpcstatus.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

func normalizeLogger(logger *log.Logger) *log.Logger {
	if logger != nil {
		return logger
	}

	return log.Default()
}
