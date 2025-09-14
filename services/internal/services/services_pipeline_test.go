package services

import (
	"pipecraft/internal/models"
	"pipecraft/internal/storage"
	"testing"

	"github.com/stretchr/testify/require"
)

type Suite struct {
	pipelineService *PipelineService
}

func NewSuite() *Suite {
	s := NewStorageMock()
	p := NewPipelineService(s)
	return &Suite{pipelineService: p}
}

func Test_PipelineService_Run_HappyPath(t *testing.T) {
	s := NewSuite()

	requestDto := models.RunPipelineRequest{
		RepositoryUrl: "repo",
		Branch:        "branch",
		Commit:        "commit",
	}

	resp, err := s.pipelineService.Run(&requestDto)
	require.NoError(t, err)
	require.Equal(t, resp.PipelineId, int64(1))

	// creating duplicate
	resp, err = s.pipelineService.Run(&requestDto)
	require.ErrorIs(t, err, ErrAlreadyExists)
	require.NotNil(t, resp)
	require.Equal(t, resp.PipelineId, int64(1))
}

func Test_PipelineService_Status_HappyPath(t *testing.T) {
	s := NewSuite()

	requestDto := models.RunPipelineRequest{
		RepositoryUrl: "repo",
		Branch:        "branch",
		Commit:        "commit",
	}

	runResponse, err := s.pipelineService.Run(&requestDto)
	require.NoError(t, err)
	require.Equal(t, runResponse.PipelineId, int64(1))

	statusResponse, err := s.pipelineService.GetStatus(runResponse.PipelineId)
	require.NoError(t, err)
	require.Equal(t, statusResponse.Status, storage.PIPELINE_STATUS_WAITING)

	statusResponse, err = s.pipelineService.GetStatus(int64(-1))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, statusResponse)
}

func Test_PipelineService_Logs_HappyPath(t *testing.T) {
	s := NewSuite()

	requestDto := models.RunPipelineRequest{
		RepositoryUrl: "repo",
		Branch:        "branch",
		Commit:        "commit",
	}

	runResponse, err := s.pipelineService.Run(&requestDto)
	require.NoError(t, err)
	require.Equal(t, runResponse.PipelineId, int64(1))

	logsResponse, err := s.pipelineService.GetLogs(runResponse.PipelineId)
	require.NoError(t, err)
	require.NotNil(t, logsResponse)

	logsResponse, err = s.pipelineService.GetLogs(int64(-1))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotFound)
	require.Nil(t, logsResponse)
}

func Test_PipelineService_Error(t *testing.T) {
	s := NewErrorStorageMock()
	p := NewPipelineService(s)

	requestDto := models.RunPipelineRequest{
		RepositoryUrl: "repo",
		Branch:        "branch",
		Commit:        "commit",
	}

	runResponse, err := p.Run(&requestDto)
	require.Error(t, err)
	require.Nil(t, runResponse)

	logsResponse, err := p.GetLogs(int64(1))
	require.Error(t, err)
	require.Nil(t, logsResponse)

	statusResponse, err := p.GetStatus(int64(1))
	require.Error(t, err)
	require.Nil(t, statusResponse)
}
