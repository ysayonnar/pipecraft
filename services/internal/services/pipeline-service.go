package services

import "pipecraft/internal/storage"

type PipelineService struct {
	Storage *storage.Storage
}

func NewPipelineService(s *storage.Storage) *PipelineService {
	return &PipelineService{Storage: s}
}
