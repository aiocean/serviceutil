package interceptor

import (
	"github.com/google/wire"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var WireSet = wire.NewSet(
	NewStreamServerInterceptor,
	NewUnaryServerInterceptor,
)

func NewStreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return grpcmiddleware.ChainStreamServer(
		grpczap.StreamServerInterceptor(logger),
		grpc_validator.StreamServerInterceptor(),
	)
}

func NewUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return grpcmiddleware.ChainUnaryServer(
		grpczap.UnaryServerInterceptor(logger),
		grpc_validator.UnaryServerInterceptor(),
	)
}
