package interceptor

import (
	"context"

	"github.com/google/wire"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

func newMsgProducer(logger *zap.Logger) grpczap.MessageProducer {
	return func(ctx context.Context, msg string, level zapcore.Level, code codes.Code, err error, duration zapcore.Field) {
		if e := logger.Check(level, msg); e != nil {
			ctxzap.Extract(ctx).Core().Write(e.Entry, []zapcore.Field{
				zap.Error(err),
				zap.String("grpc.code", code.String()),
				duration,
			})
		}
	}
}

func DefaultStreamServerInterceptor(logger *zap.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpczap.StreamServerInterceptor(logger, grpczap.WithMessageProducer(newMsgProducer(logger))),
		grpc_recovery.StreamServerInterceptor(),
		grpc_validator.StreamServerInterceptor(),
		otelgrpc.StreamServerInterceptor(),
	}

}

func DefaultUnaryServerInterceptor(logger *zap.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpczap.UnaryServerInterceptor(logger, grpczap.WithMessageProducer(newMsgProducer(logger))),
		grpc_recovery.UnaryServerInterceptor(),
		grpc_validator.UnaryServerInterceptor(),
		otelgrpc.UnaryServerInterceptor(),
	}
}
