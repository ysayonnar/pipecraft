package services

import (
	"pipecraft/internal/models"
	"pipecraft/internal/storage"
)

type PipelineService struct {
	Storage *storage.Storage
}

func NewPipelineService(s *storage.Storage) *PipelineService {
	return &PipelineService{Storage: s}
}

func (s *PipelineService) GetPipelineStatus(id int64) (*models.PipelineStatusResponse, error) {
	//TODO: implement
	return &models.PipelineStatusResponse{}, nil
}

func (s *PipelineService) GetPipelineLogs(id int64) (*models.PipelineLogsResponse, error) {
	//TODO: implement
	return &models.PipelineLogsResponse{}, nil
}
