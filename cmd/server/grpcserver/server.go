package grpcserver

import (
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	unaryMiddlewares = []grpc.UnaryServerInterceptor{}
)

func fetchKey(m metadata.MD, key string) (string, error) {
	switch vals := m.Get(key); len(vals) {
	case 0:
		return "", nil
	case 1:
		return vals[0], nil
	default:
		return "", status.Errorf(codes.InvalidArgument, "metadata value '%s' should only have one value, but got %d: %v", key, len(vals), vals)
	}
}

type server struct {
	logger *slog.Logger
	tracer trace.Tracer

	reader *sqlx.DB
	writer *sqlx.DB
}

// NewServer creates a new server. Pass an empty string to traceName to not add any trace middleware
func NewServer(traceName string, l *slog.Logger, reader, writer *sqlx.DB) *grpc.Server {
	s := &server{
		logger: l,
		tracer: otel.Tracer(traceName),
		reader: reader,
		writer: writer,
	}

	// only if tracing was specified should you add this
	if traceName != "" {
		unaryMiddlewares = append(unaryMiddlewares, otelgrpc.UnaryServerInterceptor())
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Second,
			MaxConnectionAge:  5 * time.Minute,
			Timeout:           20 * time.Second,
		}),

		grpc.ChainUnaryInterceptor(unaryMiddlewares...),

		// Stream requests
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	// add server implementations below

	reflection.Register(grpcServer)
	return grpcServer
}
