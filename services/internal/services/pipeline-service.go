package services

import (
	"errors"
	"fmt"
	"pipecraft/internal/models"
	"pipecraft/internal/storage"
)

var (
	ErrNotFound      = errors.New("pipeline not found")
	ErrAlreadyExists = errors.New("pipeline already exists")
)

type PipelineService struct {
	Storage *storage.Storage
}

func NewPipelineService(s *storage.Storage) *PipelineService {
	return &PipelineService{Storage: s}
}

func (s *PipelineService) Run(dto *models.RunPipelineRequest) (*models.RunPipelineResponse, error) {
	const op = `service.PipelineService.Run`

	pipelineId, err := s.Storage.CreatePipeline(dto.RepositoryUrl, dto.Branch, dto.Commit)
	if err != nil {
		if errors.Is(err, storage.ErrPipelineAlreadyExists) {
			return &models.RunPipelineResponse{PipelineId: pipelineId}, ErrAlreadyExists
		}
		return nil, fmt.Errorf(`%s: %w`, err, op)
	}

	return &models.RunPipelineResponse{PipelineId: pipelineId}, nil
}

func (s *PipelineService) GetStatus(id int64) (*models.PipelineStatusResponse, error) {
	const op = `services.PipelineService.GetStatus`

	status, err := s.Storage.GetPipelineStatus(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return &models.PipelineStatusResponse{Status: status}, nil
}

func (s *PipelineService) GetLogs(id int64) (*models.PipelineLogsResponse, error) {
	const op = `services.PipelineService.GetLogs`

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
