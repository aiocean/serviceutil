package tracer

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

type TracerConfig struct {
	JaegerHost        string
	JaegerPort        string
	Logger            *zap.Logger
	ServiceName       string
	ServiceInstanceId string
	ServiceNamespace  string
	ServiceVersion    string
}

var DefaultTracerSet = wire.NewSet(
	NewTracer,
)

var EnvConfigSet = wire.NewSet(
	NewConfigFromEnv,
)

func NewConfigFromEnv() (*TracerConfig, error) {
	cfg := &TracerConfig{}
	ok := false

	cfg.JaegerHost, ok = os.LookupEnv("JAEGER_HOST")
	if !ok {
		return nil, errors.New("JAEGER_HOST is not set")
	}
	cfg.JaegerPort, ok = os.LookupEnv("JAEGER_PORT")
	if !ok {
		return nil, errors.New("JAEGER_PORT is not set")
	}

	cfg.ServiceName, ok = os.LookupEnv("SERVICE_NAME")
	if !ok {
		return nil, errors.New("SERVICE_NAME is not set")
	}

	cfg.ServiceInstanceId, ok = os.LookupEnv("SERVICE_INSTANCE_ID")
	if !ok {
		return nil, errors.New("SERVICE_INSTANCE_ID is not set")
	}

	cfg.ServiceNamespace, ok = os.LookupEnv("SERVICE_NAMESPACE")
	if !ok {
		return nil, errors.New("SERVICE_NAMESPACE is not set")
	}

	cfg.ServiceVersion, ok = os.LookupEnv("SERVICE_VERSION")
	if !ok {
		return nil, errors.New("SERVICE_VERSION is not set")
	}

	return cfg, nil
}

func NewTracer(ctx context.Context, cfg *TracerConfig, logger *zap.Logger) (*tracesdk.TracerProvider, func(), error) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint())
	if err != nil {
		return nil, nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
	)

	cleanup := func() {
		// Do not make the application hang when it is shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal("failed to shutdown", zap.Error(err))
		}
	}

	return tp, cleanup, nil
}
