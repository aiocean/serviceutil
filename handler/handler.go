package handler

import (
	"context"
	"net"
	"os"

	"github.com/google/wire"
	"pkg.aiocean.dev/serviceutil/healthserver"

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
	streamServerInterceptor grpc.StreamServerInterceptor,
	unaryServerInterceptor grpc.UnaryServerInterceptor,
	healthServer *healthserver.Server,
) *Handler {

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(streamServerInterceptor),
		grpc.UnaryInterceptor(unaryServerInterceptor),
	)
	healthServer.Register(grpcServer)
	reflection.Register(grpcServer)
	serviceServer.Register(grpcServer)

	handler := &Handler{
		Address:    ":" + os.Getenv("PORT"),
		Logger:     logger,
		GrpcServer: grpcServer,
	}

	if address, hasBaseDomain := os.LookupEnv("ADDRESS"); hasBaseDomain {
		handler.Address = address
	}

	return handler
}

func (h *Handler) Serve() {
	defer h.Logger.Sync()

	h.Logger.Info("serve at: " + h.Address)

	listen, err := net.Listen("tcp", h.Address)
	if err != nil {
		h.Logger.Fatal("failed to listen", zap.Error(err))
	}

	if err := h.GrpcServer.Serve(listen); err != nil {
		h.Logger.Fatal("Failed to serve", zap.Error(err))
	}
}
