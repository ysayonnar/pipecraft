//go:build test

package handlers

import "pipecraft/internal/models"

type MockPipelineService struct{}

func (m MockPipelineService) Run(dto *models.RunPipelineRequest) (*models.RunPipelineResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockPipelineService) GetStatus(id int64) (*models.PipelineStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockPipelineService) GetLogs(id int64) (*models.PipelineLogsResponse, error) {
	//TODO implement me
	panic("implement me")
}
