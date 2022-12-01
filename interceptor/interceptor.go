package interceptor

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/google/wire"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func getZapStackTrace(panic interface{}) []zapcore.Field {
	fields := []zap.Field{
		zap.Any("panic", panic),
	}
	s := string(debug.Stack())
	stackLine := 0
	lines := strings.Split(s, "\n")
	// for line in reverse order
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.Contains(line, ".go:") {
			if !strings.Contains(line, "/vendor/") &&
				!strings.HasPrefix(line, "\t/usr/local/go/src/runtime/") {
				stackLine++

				// remove +0x... at the end of the line
				split := strings.LastIndex(line, " +0x")
				if split > 0 {
					line = line[:split]
				}

				f := zap.String(fmt.Sprintf("stack_%02d", stackLine), line)
				fields = append(fields, f)
			}
		}
	}

	return fields
}

func makeRecoveryLog(logger *zap.Logger) grpc_recovery.Option {
	return grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
		logger.Error("panic recovered", getZapStackTrace(p)...)
		return status.Errorf(codes.Internal, "%v", p)
	})
}

func DefaultStreamServerInterceptor(logger *zap.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpczap.StreamServerInterceptor(logger),
		grpc_recovery.StreamServerInterceptor((makeRecoveryLog(logger))),
		grpc_validator.StreamServerInterceptor(),
		otelgrpc.StreamServerInterceptor(),
	}

}

func DefaultUnaryServerInterceptor(logger *zap.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpczap.UnaryServerInterceptor(logger),
		grpc_recovery.UnaryServerInterceptor(makeRecoveryLog(logger)),
		grpc_validator.UnaryServerInterceptor(),
		otelgrpc.UnaryServerInterceptor(),
	}
}
