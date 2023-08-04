package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const appName = "backend"
const asciiArt = ``

// CLI vars
var (
	envStr = flag.String("env", "local", "which env: local | dev | stage | prod")

	// logging
	logExporter = flag.String("log-exporter", "", "File to log to. Blank for stdout")
	logLevel    = flag.String("log-level", "INFO", "Log level to use: DEBUG | INFO | WARN | ERROR")
	logFmt      = flag.String("log-format", "json", "Log format to use: json | logfmt")

	// metrics
	disableMetricsServer = flag.Bool("disable-metrics-server", false, "Disable the metrics server")
	httpMetricsPort      = flag.Int("http-metrics-port", 8088, "HTTP metrics port")

	// tracing
	traceExporter        = flag.String("trace-exporter", "stdout", "Export trace data. Options: stdout | jaeger | otlp | none")
	traceExporterURL     = flag.String("trace-export-url", "", "URL to use for your trace export option")
	traceExporterTimeout = flag.Duration("trace-exporter-timeout", time.Second*5, "How long the tracer will try to export before it abandons the whole process (not supported for all trace exporters)")

	// health server
	disableHealthServer = flag.Bool("disable-health", false, "disable the health check server")
	healthPort          = flag.Uint("health-port", 7674, "the port to run liveliness/readiness")

	// gRPC server
	grpcPort = flag.Uint("grpc-port", 9200, "run the GRPC server at this port")

	// gRPC gateway interface to expose HTTP
	grpcGatewayPort = flag.Uint("grpc-gateway-port", 0, "run the grpc-gateway server and listen to this port. If 0, don't use it. gRPC must be enabled for it to work")

	// db plaintext config
	dbHost = flag.String("db-host", "localhost", "the database host to connect to. If localhost, sslmode=disable; for any other host, sslmode=require")
	dbPort = flag.Uint("db-port", 5432, "what port to connect to the DB on")
	dbName = flag.String("db-name", "aq", "what database to connect to")

	// db reader user
	dbReaderUser     = flag.String("db-reader-user", "dbreader", "The database reader username")
	dbReaderPassword = flag.String("db-reader-password", "", "database reader's password")

	// db writer user
	dbWriterUser     = flag.String("db-writer-user", "dbwriter", "The database writer username")
	dbWriterPassword = flag.String("db-writer-password", "", "database writer's password")
)

// runtime vars
var (
	logger *slog.Logger
	tp     *trace.TracerProvider

	// database
	dbReader *sqlx.DB
	dbWriter *sqlx.DB

	// servers
	httpServer        *http.Server
	httpMetricsServer *http.Server
	grpcServer        *grpc.Server
	grpcGatewayServer *http.Server
	healthServer      *grpc.Server
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bootstrap(ctx) // set all important global vars

	// catch termination
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(serveGRPC)

	if *grpcGatewayPort != 0 {
		g.Go(serveGRPCGateway(ctx))
	}

	if !*disableMetricsServer {
		g.Go(prometheusMetrics)
	}

	if !*disableHealthServer {
		g.Go(serveHealth)
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		logger.Info("Starting app",
			"version", info.Main.Version,
			"path", info.Main.Path,
			"checksum", info.Main.Sum,
		)
	}

	fmt.Println(asciiArt)

	select {
	case <-interrupt:
		logger.Info("caught kill signal, canceling context and exiting")
		cancel()
	case <-ctx.Done():
		logger.Info("context canceled", "err", ctx.Err())
	}

	shutdownDeadline, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	shutdownServers(shutdownDeadline)

	if err := g.Wait(); err != nil {
		logger.Info("shutdown error", "err", err)
		os.Exit(1)
	}
}
