package cmdline

import (
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// ConnectDB opens a database, pings it, and returns it if all succeeds.
// Will determine if sslmode is needed based on the host you're connecting to
func (a *App) ConnectDB(host, name, user, password string, port uint16) (*sqlx.DB, error) {
	sslmode := useSSL(host)
	a.logger.Info("Connecting database with no middleware",
		"host", host,
		"port", port,
		"sslmode", sslmode,
		"name", name,
		"user", user,
		"len(password) > 0", len(password) > 0,
	)

	db, err := sqlx.Open("postgres", genConnStr(port, host, name, user, password, sslmode))
	if err != nil {
		a.logger.Error("msg", "failed connecting to database")
		return nil, err
	}

	if err = a.ping(db); err != nil {
		return nil, err
	}

	return db, nil
}

// ConnectDBWithOTEL will open a DB connection with OTEL middleware wrapped around it for getting logs
// on queries
func (a *App) ConnectDBWithOTEL(host, name, user, password string, port uint16) (*sqlx.DB, error) {
	sslmode := useSSL(host)
	a.logger.Info("Connecting database with OTEL middleware",
		"host", host,
		"port", port,
		"sslmode", sslmode,
		"name", name,
		"user", user,
		"password_set", len(password) > 0,
	)

	otelDB, err := otelsql.Open("postgres", genConnStr(port, host, name, user, password, sslmode), otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
	))

	if err != nil {
		a.logger.Error("failed connecting to database", "err", err)
		return nil, err
	}

	db := sqlx.NewDb(otelDB, "postgres")
	if err = a.ping(db); err != nil {
		return nil, err
	}

	return db, nil
}

func useSSL(host string) string {
	if host == "localhost" || host == "127.0.0.1" {
		return "disable" // sslmode unnecessary if localhost or 127.0.0.1
	}

	return "require"
}

func genConnStr(port uint16, host, name, user, password, useSSL string) string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%v sslmode=%v user=%s password=%s",
		host,
		port,
		name,
		useSSL,
		user,
		password,
	)
}

func (a *App) ping(db *sqlx.DB) error {
	if err := db.Ping(); err != nil {
		a.logger.Error("failed ping", "err", err)
		return err
	}

	a.logger.Info("connected and pinged database")
	return nil
}
