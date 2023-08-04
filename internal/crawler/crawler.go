package crawler

import (
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

type Crawler struct {
	logger *slog.Logger
	tracer trace.Tracer

	urls       []*url.URL
	httpClient *http.Client
}

func New(traceName string, logger *slog.Logger, client *http.Client) *Crawler {
	if client == nil {
		client = http.DefaultClient
	}

	return &Crawler{
		logger:     logger,
		tracer:     otel.Tracer(traceName),
		httpClient: client,
	}
}
