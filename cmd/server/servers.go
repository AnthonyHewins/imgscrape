package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/AnthonyHewins/imgscrape/cmd/server/grpcserver"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	versionGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "apiserver",
		Name:      "version",
		Help:      "App version",
	}, []string{"version"})
)

func prometheusMetrics() error {
	listenAddr := fmt.Sprintf(":%d", *httpMetricsPort)
	httpMetricsServer = &http.Server{
		Addr:              listenAddr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	http.Handle("/metrics", promhttp.Handler())

	if info, ok := debug.ReadBuildInfo(); ok {
		// expose version to prometheus
		versionGauge.WithLabelValues(info.Main.Version).Add(1)

		http.HandleFunc("/version", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			b, _ := json.MarshalIndent(info, "", " ")
			w.Write(b)
		})

		http.HandleFunc("/healthz", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})
	}

	logger.Info("HTTP Metrics server listening", "listenaddr", listenAddr)
	if err := httpMetricsServer.ListenAndServe(); err != http.ErrServerClosed {
		return err //nolint:wrapcheck
	}

	return nil
}

func serveGRPC() error {
	addr := fmt.Sprintf(":%d", *grpcPort)
	tcpSocket, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("gRPC server: failed to listen", "error", err)
		return err
	}

	traceName := "grpc-server"
	if *traceExporter == "" {
		traceName = ""
	}

	grpcServer = grpcserver.NewServer(traceName, logger, dbReader, dbWriter)
	logger.Info(fmt.Sprintf("gRPC API server listening at :%d", *grpcPort))
	return grpcServer.Serve(tcpSocket)
}

func serveGRPCGateway(ctx context.Context) func() error {
	// wrapper type for binding services
	type svcHandler struct {
		name       string
		svcHandler func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
	}

	return func() error {
		mux := runtime.NewServeMux()

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()), // proxy has no need for credentials
		}

		listenAddr := fmt.Sprintf(":%d", *grpcPort)
		for _, handler := range []svcHandler{} {
			err := handler.svcHandler(ctx, mux, listenAddr, opts)
			if err != nil {
				return fmt.Errorf("failed registering grpc gateway service %s: %w", handler.name, err)
			}
		}

		grpcGatewayServer = &http.Server{Addr: fmt.Sprintf(":%d", *grpcGatewayPort), Handler: mux}
		logger.Info("serving gRPC gateway", "listenAddr", listenAddr)
		return grpcGatewayServer.ListenAndServe()
	}
}

func serveHealth() error {
	h := health.NewServer()
	healthServer = grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(healthServer, h)

	haddr := fmt.Sprintf(":%d", *healthPort)
	hln, err := net.Listen("tcp", haddr)
	if err != nil {
		logger.Error("gRPC Health server: failed to listen", "error", err)
		return err
	}

	logger.Info("gRPC health server serving", "listenaddr", haddr)

	h.SetServingStatus("backend", grpc_health_v1.HealthCheckResponse_SERVING)
	return healthServer.Serve(hln)
}
