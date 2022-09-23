package handler

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/aiocean/serviceutil/healthserver"
	"github.com/aiocean/serviceutil/interceptor"
	"github.com/google/wire"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Handler struct {
	Address    string
	Logger     *zap.Logger
	GrpcServer *grpc.Server
}

var WireSet = wire.NewSet(
	NewHandler,
)

type ServiceServer interface {
	Register(grpcServer *grpc.Server)
}

func NewHandler(
	ctx context.Context,
	logger *zap.Logger,
	serviceServer ServiceServer,
	healthServer *healthserver.Server,
	interceptor *interceptor.Interceptor,
	tracerSvc *tracesdk.TracerProvider,
) *Handler {
	otel.SetTracerProvider(tracerSvc)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			otelgrpc.StreamServerInterceptor(),
			interceptor.StreamServerInterceptor,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			otelgrpc.UnaryServerInterceptor(),
			interceptor.UnaryServerInterceptor,
		)),
	)

	healthServer.Register(grpcServer)
	reflection.Register(grpcServer)
	serviceServer.Register(grpcServer)

	port := "8080"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}

	handler := &Handler{
		Address:    net.JoinHostPort("", port),
		Logger:     logger,
		GrpcServer: grpcServer,
	}

	if address, hasBaseDomain := os.LookupEnv("ADDRESS"); hasBaseDomain {
		handler.Address = address
	}

	return handler
}

func (h *Handler) Serve() {
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			log.Println("err:" + err.Error())
		}
	}(h.Logger)

	h.Logger.Info("serve at: " + h.Address)

	listen, err := net.Listen("tcp", h.Address)
	if err != nil {
		h.Logger.Fatal("failed to listen", zap.Error(err))
	}

	if err := h.GrpcServer.Serve(listen); err != nil {
		h.Logger.Fatal("Failed to serve", zap.Error(err))
	}
}
