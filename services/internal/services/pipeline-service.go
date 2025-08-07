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
	//TODO: implement
	return &models.PipelineLogsResponse{}, nil
}
