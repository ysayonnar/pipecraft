package storage

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"pipecraft/internal/logger"

	_ "github.com/lib/pq"
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

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("error while connecting to database", logger.Err(err))
		panic(err)
	}

	err = conn.Ping()
	if err != nil {
		slog.Error("error while pinging database connection", logger.Err(err))
		panic(err)
	}

	return &Storage{Db: conn}
}
