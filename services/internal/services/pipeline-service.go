package services

import (
	"errors"
	"fmt"
	"pipecraft/internal/models"
	"pipecraft/internal/storage"
)

var (
	ErrNotFound = errors.New("pipeline not found")
)

type PipelineService struct {
	Storage *storage.Storage
}

//just test

func NewPipelineService(s *storage.Storage) *PipelineService {
	return &PipelineService{Storage: s}
}

func (s *PipelineService) GetPipelineStatus(id int64) (*models.PipelineStatusResponse, error) {
	const op = `services.PipelineService.GetPipelineStatus`

	status, err := s.Storage.GetPipelineStatus(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return &models.PipelineStatusResponse{Status: status}, nil
}

func (s *PipelineService) GetPipelineLogs(id int64) (*models.PipelineLogsResponse, error) {
	const op = `services.PipelineService.GetPipelineLogs`

	logs, err := s.Storage.GetPipelineLogs(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	logsRequest := make([]models.Logs, len(logs))
	for i, logEntity := range logs {
		logsRequest[i] = models.Logs{
			LogsId:        logEntity.LogId,
			CommandNumber: logEntity.CommandNumber,
			CommandName:   logEntity.CommandName,
			Command:       logEntity.Command,
			Results:       logEntity.Results,
			FinalStatus:   logEntity.FinalStatus,
		}
	}

	return &models.PipelineLogsResponse{Logs: logsRequest}, nil
}
