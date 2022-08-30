package interceptor

import (
	"github.com/google/wire"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var DefaultWireSet = wire.NewSet(
	DefaultStreamServerInterceptor,
	DefaultUnaryServerInterceptor,
)

var WireSet = wire.NewSet(
	NewInterceptor,
)

type Interceptor struct {
	StreamServerInterceptor grpc.StreamServerInterceptor
	UnaryServerInterceptor  grpc.UnaryServerInterceptor
}

func NewInterceptor(streamServerInterceptors []grpc.StreamServerInterceptor, unaryServerInterceptors []grpc.UnaryServerInterceptor) *Interceptor {
	interceptor := Interceptor{
		StreamServerInterceptor: grpc_middleware.ChainStreamServer(streamServerInterceptors...),
		UnaryServerInterceptor:  grpc_middleware.ChainUnaryServer(unaryServerInterceptors...),
	}

	return &interceptor
}

func DefaultStreamServerInterceptor(logger *zap.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpczap.StreamServerInterceptor(logger),
		grpc_recovery.StreamServerInterceptor(),
		grpc_validator.StreamServerInterceptor(),
		otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(otel.GetTracerProvider())),
	}

}

func DefaultUnaryServerInterceptor(logger *zap.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpczap.UnaryServerInterceptor(logger),
		grpc_recovery.UnaryServerInterceptor(),
		grpc_validator.UnaryServerInterceptor(),
		otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(otel.GetTracerProvider())),
	}
}
