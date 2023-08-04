package iiif

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

type Client struct {
	logger     *slog.Logger
	tracer     trace.Tracer
	httpClient *http.Client
	baseURL    string
}

func NewClient(traceName string, logger *slog.Logger, httpClient *http.Client, baseURL string) *Client {
	return &Client{
		logger:     logger,
		tracer:     otel.Tracer(traceName),
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}
