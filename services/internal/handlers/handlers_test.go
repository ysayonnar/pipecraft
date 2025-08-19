package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pipecraft/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlers_RunPipeline_HappyPath(t *testing.T) {
	redisMock := NewMockRedisServie()
	pipelinesMock := NewMockPipelineService()
	handlers := New(redisMock, pipelinesMock)

	pipeline := models.RunPipelineRequest{
		RepositoryUrl: "ysayonnar/pipecraft",
		Branch:        "main",
		Commit:        "e4r3e2",
	}

	requestBody, _ := json.Marshal(pipeline)
	req, _ := http.NewRequest(http.MethodPost, "/run-pipeline", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var response models.RunPipelineResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, response.PipelineId, int64(1))

	// checking if handler detects similar pipeline
	req, _ = http.NewRequest(http.MethodPost, "/run-pipeline", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, response.PipelineId, int64(1))
}

func TestHandlers_RunPipeline_MethodNotAllowed(t *testing.T) {
	redisMock := NewMockRedisServie()
	pipelinesMock := NewMockPipelineService()
	handlers := New(redisMock, pipelinesMock)

	req, _ := http.NewRequest(http.MethodGet, "/run-pipeline", nil)
	rr := httptest.NewRecorder()

	handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
