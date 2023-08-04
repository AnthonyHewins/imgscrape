package main

import (
	"context"

	"github.com/AnthonyHewins/csvscan"
	"golang.org/x/exp/slog"
)

type row struct {
	ID                 string `csv:"0"`
	ImageURL           string `csv:"1"`
	ThumbURL           string `csv:"2"`
	ViewType           string `csv:"3"`
	Sequence           string `csv:"4"`
	Width              string `csv:"5"`
	Height             string `csv:"6"`
	MaxPixels          string `csv:"7"`
	Created            string `csv:"8"`
	Modified           string `csv:"9"`
	DepictSTMSObjectID string `csv:"10"`
	AssistiveText      string `csv:"11"`
}

func csv(ctx context.Context, logger *slog.Logger, filename string) ([]row, error) {
	reader := csvscan.Reader[row]{IgnoreHeader: true}

	l := logger.With("filename", filename)
	l.InfoContext(ctx, "reading from file")
	rows, err := reader.ReadFile(filename)
	if err != nil {
		logger.ErrorContext(ctx, "failed reading file", "err", err)
		return nil, err
	}

	return rows, nil
}
