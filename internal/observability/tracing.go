package observability

import (
	"context"
	"fmt"

	"aipivot/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracing(ctx context.Context, conf config.TelemetryConf) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpointURL(conf.JaegerEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	sampler := sdktrace.AlwaysSample()
	if conf.SampleRatio >= 0 && conf.SampleRatio < 1 {
		sampler = sdktrace.TraceIDRatioBased(conf.SampleRatio)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes(
			"",
			attribute.String("service.name", conf.ServiceName),
			attribute.String("deployment.environment", conf.Environment),
		)),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logx.Errorf("otel error: %v", err)
	}))

	logx.Infof("tracing initialized: endpoint=%s service=%s env=%s ratio=%.2f",
		conf.JaegerEndpoint, conf.ServiceName, conf.Environment, conf.SampleRatio)

	return provider.Shutdown, nil
}
