package cmdline

import (
	"net/http"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type App struct {
	appName    string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewAppFromCobra(appName string, cmd *cobra.Command) (*App, error) {
	f := cmd.Flags()

	logger, err := getLoggerFromCobra(appName, f)
	if err != nil {
		return nil, err
	}

	timeout, err := f.GetDuration("http-timeout")
	if err != nil {
		return nil, err
	}

	return &App{
		appName:    appName,
		logger:     logger,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func NewApp(appName, logLevel, logFmt, exporter string, addSrc bool) (*App, error) {
	logger, err := getLogger(appName, logLevel, logFmt, exporter, addSrc)
	if err != nil {
		return nil, err
	}

	return &App{
		appName: appName,
		logger:  logger,
	}, nil
}

func (a *App) Logger() *slog.Logger {
	return a.logger
}
