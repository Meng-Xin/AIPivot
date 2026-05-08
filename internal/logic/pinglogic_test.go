package logic

import (
	"context"
	"testing"

	"aipivot/internal/observability"
	"aipivot/internal/svc"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestPingReturnsMessageAndRequestID(t *testing.T) {
	ctx := observability.ContextWithRequestID(context.Background(), "request-123")
	logic := NewPingLogic(ctx, &svc.ServiceContext{})

	resp, err := logic.Ping()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Message != "pong" {
		t.Fatalf("expected pong, got %q", resp.Message)
	}
	if resp.RequestID != "request-123" {
		t.Fatalf("expected request-123, got %q", resp.RequestID)
	}
}

func TestPingReturnsTraceID(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	tracer := provider.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	ctx = observability.ContextWithRequestID(ctx, "request-123")
	logic := NewPingLogic(ctx, &svc.ServiceContext{})

	resp, err := logic.Ping()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.TraceID == "" {
		t.Fatalf("expected trace id")
	}
}
