package interceptor

import (
	"github.com/google/wire"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
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
		grpcrecovery.StreamServerInterceptor(),
		grpc_validator.StreamServerInterceptor(),
	)
}

func NewUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return grpcmiddleware.ChainUnaryServer(
		grpczap.UnaryServerInterceptor(logger),
		grpcrecovery.UnaryServerInterceptor(),
		grpc_validator.UnaryServerInterceptor(),
	)
}
