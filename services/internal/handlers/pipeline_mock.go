package handlers

import (
	"pipecraft/internal/models"
	"pipecraft/internal/services"
	"pipecraft/internal/storage"
	"time"
)

type MockPipelineService struct {
	pipelines      map[int64]*storage.PipelinesTable
	logs           map[int64]*storage.LogsTable
	lastPipelineId int64
	lastLogId      int64
}

func NewMockPipelineService() *MockPipelineService {
	return &MockPipelineService{
		pipelines:      make(map[int64]*storage.PipelinesTable),
		logs:           make(map[int64]*storage.LogsTable),
		lastPipelineId: 0,
		lastLogId:      0,
	}
}

func (m MockPipelineService) Run(dto *models.RunPipelineRequest) (*models.RunPipelineResponse, error) {
	for _, pipeline := range m.pipelines {
		if pipeline.Repository == dto.RepositoryUrl && pipeline.Commit == dto.Commit && pipeline.Branch == dto.Branch {
			return &models.RunPipelineResponse{PipelineId: pipeline.PipelineId}, services.ErrAlreadyExists
		}
	}

	m.lastPipelineId++
	m.lastLogId++

	m.pipelines[m.lastPipelineId] = &storage.PipelinesTable{
		PipelineId: m.lastPipelineId,
		Status:     storage.PIPELINE_STATUS_WAITING,
		Repository: dto.RepositoryUrl,
		Branch:     dto.Branch,
		Commit:     dto.Commit,
		CreatedAt:  time.Now(),
	}

	m.logs[m.lastLogId] = &storage.LogsTable{
		LogId:         m.lastLogId,
		CommandNumber: 1,
		CommandName:   "build",
		Command:       "docker build name-of-dockerfile",
		Results:       "built",
		FinalStatus:   "succeeded",
		PipelineId:    m.lastPipelineId,
	}

	return &models.RunPipelineResponse{PipelineId: m.lastPipelineId}, nil
}

func (m MockPipelineService) GetStatus(id int64) (*models.PipelineStatusResponse, error) {
	pipeline, ok := m.pipelines[id]
	if !ok {
		return nil, services.ErrNotFound
	}

	return &models.PipelineStatusResponse{Status: pipeline.Status}, nil
}

func (m MockPipelineService) GetLogs(id int64) (*models.PipelineLogsResponse, error) {
	pipeline, ok := m.pipelines[id]
	if !ok {
		return nil, services.ErrNotFound
	}

	for _, logs := range m.logs {
		if logs.PipelineId == pipeline.PipelineId {
			return &models.PipelineLogsResponse{Logs: []models.Logs{
				models.Logs{
					LogsId:        logs.LogId,
					CommandNumber: logs.CommandNumber,
					CommandName:   logs.CommandName,
					Command:       logs.Command,
					Results:       logs.Results,
					FinalStatus:   logs.FinalStatus,
				},
			}}, nil
		}
	}

	return nil, services.ErrNotFound
}
