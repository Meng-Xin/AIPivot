package infra

import (
	"context"
	"errors"
	"testing"
)

func TestCheckDependenciesReturnsReadyWhenAllChecksPass(t *testing.T) {
	checks := []DependencyCheck{
		{Name: "postgres", Check: func(context.Context) error { return nil }},
		{Name: "redis", Check: func(context.Context) error { return nil }},
	}

	result := CheckDependencies(context.Background(), checks)

	if !result.Ready {
		t.Fatalf("expected result to be ready")
	}
	if result.Status != "ready" {
		t.Fatalf("expected status ready, got %q", result.Status)
	}
	if len(result.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(result.Dependencies))
	}

	for _, dep := range result.Dependencies {
		if !dep.Ready {
			t.Fatalf("expected dependency %s to be ready", dep.Name)
		}
		if dep.Error != "" {
			t.Fatalf("expected dependency %s to have empty error, got %q", dep.Name, dep.Error)
		}
	}
}

func TestCheckDependenciesReturnsNotReadyWhenAnyCheckFails(t *testing.T) {
	checks := []DependencyCheck{
		{Name: "postgres", Check: func(context.Context) error { return nil }},
		{Name: "redis", Check: func(context.Context) error { return errors.New("connection refused") }},
	}

	result := CheckDependencies(context.Background(), checks)

	if result.Ready {
		t.Fatalf("expected result to be not ready")
	}
	if result.Status != "not_ready" {
		t.Fatalf("expected status not_ready, got %q", result.Status)
	}
	if result.Dependencies[1].Ready {
		t.Fatalf("expected redis to be not ready")
	}
	if result.Dependencies[1].Error != "connection refused" {
		t.Fatalf("expected redis error, got %q", result.Dependencies[1].Error)
	}
}
