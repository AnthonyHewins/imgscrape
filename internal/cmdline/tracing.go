package cmdline

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// MustCreateTracer creates a tracer to use for checking performance. Pass the app's context, a format, and a URI-like
// string. The format can be one of the following:
//
//	"stdout"   // STDOUT style export: if you pass a non-empty string for the URI, it will be redirected to a file
//	"otlp"     // OTLP style export: export to an OTLP endpoint
//	"jaegar"   // Jaegar export: export to a jaegar endpoint
//	"none"|""  // Create a no-op tracer
//
// Anything else will error out
func (a *App) CreateTracer(ctx context.Context, traceFormat, traceExportURI string, timeout time.Duration) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	switch strings.ToLower(traceFormat) {
	case "stdout":
		file := os.Stdout
		if traceExportURI != "" {
			f, err := os.Open(traceExportURI)
			if err != nil {
				a.logger.Error("failed opening the file supplied for tracing",
					"err", err,
					"path", traceExportURI,
				)
				return nil, err
			}

			file = f
		}

		exp, err := stdouttrace.New(
			stdouttrace.WithWriter(file),
			stdouttrace.WithoutTimestamps(),
		)

		if err != nil {
			a.logger.Error("can't create stdout tracer", "error", err)
			return nil, err
		}

		exporter = exp
	case "otlp":
		extraOpts := make([]otlptracegrpc.Option, 0, 1)
		if traceExportURI != "" {
			extraOpts = append(extraOpts, otlptracegrpc.WithEndpoint(traceExportURI))
		}

		exp, err := otlptracegrpc.New(ctx,
			append(extraOpts,
				otlptracegrpc.WithInsecure(),
				otlptracegrpc.WithReconnectionPeriod(time.Second),
				otlptracegrpc.WithTimeout(timeout),
			)...,
		)

		if err != nil {
			a.logger.Error("error creating otlptracegrpc exporter",
				"otlp-export-url", traceExportURI,
				"timeout", timeout,
				"err", err,
			)

			return nil, err
		}

		exporter = exp
	case "jaeger":
		extraOpts := make([]jaeger.CollectorEndpointOption, 0, 1)
		if traceExportURI != "" {
			extraOpts = append(extraOpts, jaeger.WithEndpoint(traceExportURI))
		}

		exp, err := jaeger.New(
			jaeger.WithCollectorEndpoint(extraOpts...),
		)

		if err != nil {
			a.logger.Error("can't create jaeger tracer",
				"error", err,
				"trace-export-url", traceExportURI,
			)

			return nil, err
		}

		exporter = exp
	case "none", "":
		exporter = tracetest.NewNoopExporter()
	default:
		a.logger.Error("invalid tracer", "tracer", traceFormat)
		panic("invalid tracer exporter format passed: " + traceFormat)
	}

	a.logger.Info("Creating tracer",
		"exporter", traceFormat,
		"export-url", traceExportURI,
		"exporter-type", fmt.Sprintf("%T", exporter),
	)

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(a.newResource()),
	), nil
}

// newResource returns a resource describing this application.
func (a *App) newResource() *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(a.appName),
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		attrs = append(attrs, semconv.ServiceVersionKey.String(info.Main.Version))
	}

	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)

	return r
}
