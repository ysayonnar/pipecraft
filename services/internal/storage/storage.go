package storage

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"pipecraft/internal/logger"
	"time"

	_ "github.com/lib/pq"
)

const (
	DEFAULT_CONNECTION_DELAY = time.Second
	DEFAULT_RETRIES          = 5
)

type Storage struct {
	Db *sql.DB
}

func MustInit() *Storage {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		slog.Error("empty DSN")
		panic(errors.New("Empty DSN"))
	}

	conn, err := tryToConnect(dsn)
	if err != nil {
		slog.Error("error while connecting to database", logger.Err(err))
		panic(err)
	}

	//TODO: create tables here or migrations idk

	return &Storage{Db: conn}
}

func tryToConnect(dsn string) (*sql.DB, error) {
	for r := 0; ; r++ {
		conn, err := sql.Open("postgres", dsn)
		err = conn.Ping()
		if err == nil || r >= DEFAULT_RETRIES {
			return conn, err
		}

		slog.Warn("database is unavailable to connect")

		<-time.After(DEFAULT_CONNECTION_DELAY)
	}
}
