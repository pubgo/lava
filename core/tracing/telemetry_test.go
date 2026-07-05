package tracing_test

import (
	"context"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/v2/core/tracing"
)

func TestNewProviderStdoutShutdown(t *testing.T) {
	t.Parallel()

	lc := lifecyclebuilder.New(nil)
	cfg := tracing.DefaultCfg()
	cfg.TraceExporter.ExporterEndpoint = tracing.DefaultStdout

	provider := tracing.NewProvider(&cfg, lc.Setter)
	if provider.TracerProvider == nil || provider.Tracer == nil {
		t.Fatal("expected tracer provider and tracer")
	}
	if provider.MeterProvider == nil || provider.Meter == nil {
		t.Fatal("expected meter provider and meter")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, hook := range lc.Getter.GetAfterStops() {
		if err := hook.Exec(ctx); err != nil {
			t.Fatalf("shutdown hook failed: %v", err)
		}
	}
}

func TestNewProviderDisabledIsNoop(t *testing.T) {
	t.Parallel()

	provider := tracing.NewProvider(nil, nil)
	if provider.Tracer == nil || provider.Meter == nil {
		t.Fatal("expected noop tracer and meter")
	}
}

func TestDefaultConfigDisabled(t *testing.T) {
	t.Parallel()

	cfg := tracing.DefaultCfg()
	if cfg.TraceExporter.ExporterEndpoint != "" {
		t.Fatalf("expected empty trace exporter by default, got %q", cfg.TraceExporter.ExporterEndpoint)
	}
}
