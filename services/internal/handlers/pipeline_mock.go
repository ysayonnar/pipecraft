//go:build test

package handlers

import "pipecraft/internal/models"

type MockPipelineService struct{}

func (m MockPipelineService) GetPipelineStatus(id int64) (*models.PipelineStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockPipelineService) GetPipelineLogs(id int64) (*models.PipelineLogsResponse, error) {
	//TODO implement me
	panic("implement me")
}
