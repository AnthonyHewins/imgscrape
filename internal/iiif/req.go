package iiif

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"

	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

//go:generate enumer -type quality -trimprefix quality
type quality byte

const (
	qualityDefault quality = iota
	qualityColor
	qualityGray
	qualityBitonal
)

//go:generate enumer -type format
type format byte

const (
	jpg format = iota
	tif
	png
	gif
	jp2
	pdf
	webp
)

type ImageReq struct {
	logger     *slog.Logger
	tracer     trace.Tracer
	httpClient *http.Client

	baseURL, id              string
	region, size             string
	rotation string
	quality
	format
}

func (c *Client) NewImageReq(id string) *ImageReq {
	return &ImageReq{
		logger:     c.logger,
		tracer:     c.tracer,
		httpClient: c.httpClient,
		baseURL:    c.baseURL,
		id:         id,
	}
}

func (r *ImageReq) Square() *ImageReq {
	r.region = "square"
	return r
}

func (r *ImageReq) RegionXYWH(x, y, w, h uint64) *ImageReq {
	r.region = fmt.Sprintf("%d,%d,%d,%d", x, y, w, h)
	return r
}

func (r *ImageReq) RegionPercentXYWH(x, y, w, h uint64) *ImageReq {
	r.region = fmt.Sprintf("pct:%d,%d,%d,%d", x, y, w, h)
	return r
}

func (r *ImageReq) SizeFull() *ImageReq {
	r.size = "full"
	return r
}

func (r *ImageReq) SizeMax() *ImageReq {
	r.size = "max"
	return r
}

func (r *ImageReq) SizeFixedWidthScaleHeight(w uint64) *ImageReq {
	r.size = fmt.Sprintf("%d,", w)
	return r
}

func (r *ImageReq) SizeFixedHeightScaleWidth(h uint64) *ImageReq {
	r.size = fmt.Sprintf(",%d", h)
	return r
}

func (r *ImageReq) SizePercentage(n float64) *ImageReq {
	r.size = fmt.Sprintf("pct:%f", n)
	return r
}

func (r *ImageReq) SizeWidthHeight(w, h float64) *ImageReq {
	r.size = fmt.Sprintf("%f,%f", w, h)
	return r
}

// The image content is scaled for the best fit such that the resulting width
// and height are less than or equal to the requested width and height. The
// exact scaling may be determined by the service provider, based on
// characteristics including image quality and system performance. The
// dimensions of the returned image content are calculated to maintain the
// aspect ratio of the extracted region.
func (r *ImageReq) SizeBestScaleUnder(w, h float64) *ImageReq {
	r.size = fmt.Sprintf("!%f,%f", w, h)
	return r
}

func (r *ImageReq) MirrorRotate(degrees float64) *ImageReq {
	r.rotation = fmt.Sprintf("!%f", math.Remainder(degrees, 360))
	return r
}

func (r *ImageReq) Rotate(degrees float64) *ImageReq {
	r.rotation = fmt.Sprintf("%f", math.Remainder(degrees, 360))
	return r
}

func (r *ImageReq) Jpg() *ImageReq  { r.format = jpg; return r }
func (r *ImageReq) Tif() *ImageReq  { r.format = tif; return r }
func (r *ImageReq) PNG() *ImageReq  { r.format = png; return r }
func (r *ImageReq) Gif() *ImageReq  { r.format = gif; return r }
func (r *ImageReq) Jp2() *ImageReq  { r.format = jp2; return r }
func (r *ImageReq) PDF() *ImageReq  { r.format = pdf; return r }
func (r *ImageReq) WebP() *ImageReq { r.format = webp; return r }

func (r *ImageReq) DefaultQuality() *ImageReq { r.quality = qualityDefault; return r }
func (r *ImageReq) Color() *ImageReq          { r.quality = qualityColor; return r }
func (r *ImageReq) Gray() *ImageReq           { r.quality = qualityGray; return r }
func (r *ImageReq) Bitonal() *ImageReq        { r.quality = qualityBitonal; return r }

func (r *ImageReq) Resolve(ctx context.Context) (io.Reader, error) {
	ctx, span := r.tracer.Start(ctx, "Making request to "+r.id)
	defer span.End()

	var err error
	defer func() {
		if err == nil {
			span.SetStatus(otelCodes.Ok, "Successful request")
			return
		}

		span.SetStatus(otelCodes.Error, "Failed request")
		span.RecordError(err)
	}()

	path := r.buildURL()
	l := r.logger.With("object", r, "path", path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		l.ErrorContext(ctx, "failed creating request", "err", err)
		return nil, err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		l.ErrorContext(ctx, "failed HTTP request", "err", err)
		return nil, err
	}

	switch code := resp.StatusCode {
	case 400, 401, 403, 404, 500, 501, 503:
		errMsg, err := io.ReadAll(resp.Body)
		if err != nil {
			l.ErrorContext(ctx, "failed reading err response body", "err", err)
			return nil, err
		}

		l.ErrorContext("msg", "bad response received", "code", code, "response", errMsg)
		return nil, fmt.Errorf("received %d: %s", errMsg)
	}

	if code < 200 || code >= 300 {
		l.ErrorContext(ctx, "bad response code", "err", err)
		return nil, fmt.Errorf("bad response code received: %d", code)
	}

	return resp.Body, err
}

func (r *ImageReq) readErrorResponse(resp *http.Response) error {
		errMsg, err := io.ReadAll(resp.Body)
		if err != nil {
			l.ErrorContext(ctx, "failed reading err response body", "err", err)
			return nil, err
		}

		l.ErrorContext("msg", "bad response received", "code", code, "response", errMsg)
		return nil, fmt.Errorf("received %d: %s", errMsg)
}

func (r *ImageReq) buildURL() string {
	var rotation string
	if mr := r.mirrorRotation; mr != 0 {
		rotation = fmt.Sprintf("!%f", mr)
	} else {
		rotation = fmt.Sprint(r.rotation)
	}

	return fmt.Sprintf(
		"%s/%s/%s/%s/%s/%s.%s",
		r.baseURL,
		r.id,
		r.region,
		r.size,
		rotation,
		r.quality,
		r.format,
	)
}
