package worker

import (
	"errors"
	"log/slog"
	"pipecraft/internal/logger"
	"pipecraft/internal/storage"
	"time"
)

const (
	MAX_WORKERS     = 5
	LISTEN_INTERVAL = 10
)

func StartListener(s *storage.Storage) {
	workerPool := make(chan struct{}, MAX_WORKERS)

	for {
		pipelineId, err := s.GetLastWaitingPipeline()
		if err != nil {
			if !errors.Is(err, storage.ErrNotFound) {
				slog.Error("error while getting last pipeline waiting pipeline", logger.Err(err))
			}
		}

		workerPool <- struct{}{}
		go func(pipelineId int64) {
			defer func() { <-workerPool }()
			//TODO: вызывать функцию для запуска пайплайна
		}(pipelineId)

		time.Sleep(time.Duration(LISTEN_INTERVAL) * time.Second)
	}
}
