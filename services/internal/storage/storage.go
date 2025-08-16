package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"pipecraft/internal/logger"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("not found")
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

func (s *Storage) GetPipelineStatus(id int64) (string, error) {
	const op = `storage.GetPipelineStatus`

	query := `
		SELECT
			status
		FROM 
			pipelines
		WHERE 
			pipeline_id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var status string
	err := s.Db.QueryRowContext(ctx, query, id).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("op: %s, err: %w", op, err)
	}

	return status, nil
}

func (s *Storage) GetPipelineLogs(id int64) ([]*LogsTable, error) {
	const op = `storage.GetPipelineLogs`

	query := `
		SELECT
			log_id,
			command_number,
			command_name,
			command,
			results,
			final_status
		FROM
			logs
		WHERE
			pipeline_fk_id = $1
		ORDER BY
			command_number;
    `

	logs := make([]*LogsTable, 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rows, err := s.Db.QueryContext(ctx, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	for rows.Next() {
		logEntity := &LogsTable{}

		err = rows.Scan(
			&logEntity.LogId,
			&logEntity.CommandNumber,
			&logEntity.CommandName,
			&logEntity.Command,
			&logEntity.Results,
			&logEntity.FinalStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("op: %s, err: %w", op, err)
		}

		logs = append(logs, logEntity)
	}

	return logs, nil
}
