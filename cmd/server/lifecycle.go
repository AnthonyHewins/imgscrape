package main

import (
	"context"
	"os"

	"github.com/AnthonyHewins/imgscrape/internal/cmdline"
	"github.com/namsral/flag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func bootstrap(ctx context.Context) {
	flag.Parse()

	app, err := cmdline.NewApp(appName, *logLevel, *logFmt, *logExporter, true)
	if err != nil {
		panic(err)
	}

	// logging
	logger = app.Logger()

	// database
	dbReader, err = app.ConnectDBWithOTEL(*dbHost, *dbName, *dbReaderUser, *dbReaderPassword, uint16(*dbPort))
	if err != nil {
		panic(err)
	}

	dbWriter, err = app.ConnectDBWithOTEL(*dbHost, *dbName, *dbWriterUser, *dbWriterPassword, uint16(*dbPort))
	if err != nil {
		panic(err)
	}

	// OTEL
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)

	// Tracing
	tp, err = app.CreateTracer(ctx, *traceExporter, *traceExporterURL, *traceExporterTimeout)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("failed shutting down tracer", "error", err)
			os.Exit(1)
		}
	}()
	otel.SetTracerProvider(tp) // set the global tracer provider
}

func shutdownServers(ctx context.Context) {
	if httpServer != nil {
		logger.Info("shutting down HTTP server")
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failure", "err", err)
		}
		logger.Info("HTTP server shut down")
	}

	if grpcGatewayServer != nil {
		logger.Info("shutting down gRPC gateway server")
		if err := grpcGatewayServer.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failure", "err", err)
		}
		logger.Info("grpc gateway server shut down")
	}

	if grpcServer != nil {
		logger.Info("shutting down gRPC")
		grpcServer.GracefulStop()
		logger.Info("gRPC server shut down")
	}

	if httpMetricsServer != nil {
		logger.Info("shutting down HTTP metrics")
		if err := httpMetricsServer.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failure", "err", err)
		}
		logger.Info("HTTP metrics server shut down")
	}

	if healthServer != nil {
		logger.Info("shutting down health server")
		healthServer.GracefulStop()
		logger.Info("shut down health server")
	}
}
