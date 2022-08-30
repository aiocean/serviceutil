package tracer

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap"
)

type TracerConfig struct {
	JaegerUrl         string
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

	cfg.JaegerUrl, ok = os.LookupEnv("JAEGER_URL")
	if !ok {
		return nil, errors.New("JAEGER_URL is not set")
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
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerUrl)))
	if err != nil {
		return nil, nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceInstanceIDKey.String(cfg.ServiceInstanceId),
			semconv.ServiceNamespaceKey.String(cfg.ServiceNamespace),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		)),
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
