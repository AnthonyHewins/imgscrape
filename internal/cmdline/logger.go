package cmdline

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/spf13/pflag"
	"golang.org/x/exp/slog"
)

const (
	LogSource   = "log-source"
	LogLevel    = "log-level"
	LogFmt      = "log-format"
	LogExporter = "log-exporter"
)

func getLoggerFromCobra(appName string, flags *pflag.FlagSet) (*slog.Logger, error) {
	addSrc, err := flags.GetBool(LogSource)
	if err != nil {
		return nil, err
	}

	l, err := flags.GetString(LogLevel)
	if err != nil {
		return nil, err
	}

	exporter, err := flags.GetString(LogExporter)
	if err != nil {
		return nil, err
	}

	format, err := flags.GetString(LogFmt)
	if err != nil {
		return nil, err
	}

	return getLogger(appName, l, format, exporter, addSrc)
}

func getLogger(appName, logLevel, logFmt, exporter string, addSrc bool) (*slog.Logger, error) {
	var level slog.HandlerOptions
	switch logLevel {
	case "":
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.Level(math.MaxInt)})), nil
	case "debug":
		level = slog.HandlerOptions{Level: slog.LevelDebug, AddSource: addSrc}
	case "info":
		level = slog.HandlerOptions{Level: slog.LevelInfo, AddSource: addSrc}
	case "warn":
		level = slog.HandlerOptions{Level: slog.LevelWarn, AddSource: addSrc}
	case "err":
		level = slog.HandlerOptions{Level: slog.LevelError, AddSource: addSrc}
	default:
		return nil, fmt.Errorf("invalid log level: %s", logLevel)
	}

	out, err := logExporter(exporter)
	if err != nil {
		return nil, err
	}

	var logger *slog.Logger
	switch logFmt {
	case "", "json":
		logger = slog.New(slog.NewJSONHandler(out, &level))
	case "text", "logfmt":
		logger = slog.New(slog.NewTextHandler(out, &level))
	default:
		return nil, fmt.Errorf("invalid handler format: %s", logFmt)
	}

	if appName == "" {
		logger = logger.With("app-name", appName)
	}

	return logger, nil
}

func logExporter(exporter string) (io.Writer, error) {
	switch exporter {
	case "":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	}

	file, err := os.Create(exporter)
	if err != nil {
		return nil, err
	}

	return file, nil
}
