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
	ErrNotFound              = errors.New("not found")
	ErrPipelineAlreadyExists = errors.New("pipeline already exists")
)

const (
	DEFAULT_CONNECTION_DELAY = time.Second
	DEFAULT_RETRIES          = 5

	PIPELINE_STATUS_WAITING   = "waiting"
	PIPELINE_STATUS_RUNNING   = "running"
	PIPELINE_STATUS_ABORTED   = "aborted"
	PIPELINE_STATUS_FAILED    = "failed"
	PIPELINE_STATUS_COMPLETED = "completed"
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

func (s *Storage) CreatePipeline(repository, branch, commit string) (int64, error) {
	const op = `storage.CreatePipeline`

	tx, err := s.Db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	selectQuery := `
		SELECT
			pipeline_id
		FROM
			pipelines
		WHERE
			repository = $1 AND branch = $2 AND commit = $3;
	`

	var pipelineId int64
	err = tx.QueryRow(selectQuery, repository, branch, commit).Scan(&pipelineId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		return pipelineId, ErrPipelineAlreadyExists
	}

	insertQuery := `INSERT INTO pipelines (status, repository, branch, commit) VALUES ($1, $2, $3, $4) RETURNING pipeline_id;`
	err = tx.QueryRow(insertQuery, PIPELINE_STATUS_WAITING, repository, branch, commit).Scan(&pipelineId)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return pipelineId, nil
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

func (s *Storage) GetLastWaitingPipeline() (int64, error) {
	const op = `storage.GetLastWaitingPipeline`

	query := `
		SELECT
			pipeline_id
		FROM
			pipelines
		WHERE
		    status = $1
		ORDER BY 
			created_at ASC 
		LIMIT 1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var pipelineId int64
	err := s.Db.QueryRowContext(ctx, query, PIPELINE_STATUS_WAITING).Scan(&pipelineId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return pipelineId, nil
}

func (s *Storage) UpdatePipelineStatus(id int64, status string) error {
	const op = `storage.UpdatePipelineStatus`

	query := `
		UPDATE pipelines
		SET status = $1
		WHERE pipeline_id = $2;
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := s.Db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Storage) GetPipelineInfo(id int64) (*PipelinesTable, error) {
	const op = `storage.GetPipelineInfo`

	query := `
		SELECT
			repository,
			branch,
			commit
		FROM 
			pipelines
		WHERE
			pipeline_id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	row := s.Db.QueryRowContext(ctx, query, id)
	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("op: %s, err: %w", op, ErrNotFound)
	}

	var pipeline *PipelinesTable
	if err := row.Scan(&pipeline.Repository, &pipeline.Branch, &pipeline.Commit); err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return pipeline, nil
}

func (s *Storage) CreateLog(logTable LogsTable) error {
	const op = `storage.CreateLog`

	query := `INSERT INTO logs(pipeline_fk_id, command_number, command, result, final_status) VALUES ($1, $2, $3, $4, $5);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, query, logTable.PipelineId, logTable.CommandNumber, logTable.Command, logTable.Results, logTable.FinalStatus)
	if err != nil {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	return nil
}
