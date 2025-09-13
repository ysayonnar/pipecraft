package services

import (
	"pipecraft/internal/storage"
	"time"
)

type StorageMock struct {
	pipelines      map[int64]*storage.PipelinesTable
	logs           map[int64]*storage.LogsTable
	lastPipelineId int64
	lastLogId      int64
}

func NewStorageMock() *StorageMock {
	return &StorageMock{
		pipelines:      make(map[int64]*storage.PipelinesTable),
		logs:           make(map[int64]*storage.LogsTable),
		lastPipelineId: 0,
		lastLogId:      0,
	}
}

func (s StorageMock) CreatePipeline(repository, branch, commit string) (int64, error) {
	for id, pipeline := range s.pipelines {
		if pipeline.Repository == repository && pipeline.Commit == commit && pipeline.Branch == branch {
			return id, storage.ErrPipelineAlreadyExists
		}
	}

	s.lastPipelineId++
	s.lastLogId++

	s.pipelines[s.lastPipelineId] = &storage.PipelinesTable{
		PipelineId: s.lastPipelineId,
		Status:     storage.PIPELINE_STATUS_WAITING,
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		CreatedAt:  time.Now(),
	}

	s.logs[s.lastLogId] = &storage.LogsTable{
		LogId:         s.lastLogId,
		CommandNumber: 1,
		CommandName:   "build",
		Command:       "docker build name-of-dockerfile",
		Results:       "built",
		FinalStatus:   "succeeded",
		PipelineId:    s.lastPipelineId,
	}

	return s.lastPipelineId, nil
}

func (s StorageMock) GetPipelineStatus(id int64) (string, error) {
	pipeline, ok := s.pipelines[id]
	if !ok {
		return "", storage.ErrNotFound
	}

	return pipeline.Status, nil
}

func (s StorageMock) GetPipelineLogs(id int64) ([]*storage.LogsTable, error) {
	_, ok := s.pipelines[id]
	if !ok {
		return nil, storage.ErrNotFound
	}

	for _, logs := range s.logs {
		if logs.LogId == id {
			return []*storage.LogsTable{logs}, nil
		}
	}

	return nil, storage.ErrNotFound
}
