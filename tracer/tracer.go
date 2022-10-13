package tracer

import (
	"context"
	"strings"
	"time"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var DefaultTracerSet = wire.NewSet(
	NewTracer,
)

type normalSampler struct{}

func (_ *normalSampler) ShouldSample(p tracesdk.SamplingParameters) tracesdk.SamplingResult {
	psc := trace.SpanContextFromContext(p.ParentContext)

	if strings.HasSuffix(p.Name, "Health/Check") || strings.HasSuffix(p.Name, "/healthz") {
		return tracesdk.SamplingResult{
			Decision:   tracesdk.Drop,
			Tracestate: psc.TraceState(),
		}
	}

	return tracesdk.SamplingResult{
		Decision:   tracesdk.RecordAndSample,
		Tracestate: psc.TraceState(),
	}
}

func (_ *normalSampler) Description() string {
	return "NormalSampler"
}

// NewTracer create a default tracer provider with jaeger agent
// these env are necessary for initialization
// - OTEL_EXPORTER_JAEGER_AGENT_HOST for the agent address host
// - OTEL_EXPORTER_JAEGER_AGENT_PORT for the agent address port
// - OTEL_SERVICE_NAME for service name
func NewTracer(ctx context.Context, logger *zap.Logger) (trace.TracerProvider, func(), error) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint())
	if err != nil {
		return nil, nil, err
	}
	rs, err := resource.New(ctx,
		resource.WithFromEnv(), // pull attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables
	)
	if err != nil {
		return nil, nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(&normalSampler{}),
		tracesdk.WithResource(rs),
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
