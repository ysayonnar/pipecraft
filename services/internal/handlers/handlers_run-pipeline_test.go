package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"pipecraft/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

type Suite struct {
	handlers *Handlers
}

func NewSuite() *Suite {
	redisMock := NewMockRedisServie()
	pipelinesMock := NewMockPipelineService()
	handlers := New(redisMock, pipelinesMock)
	return &Suite{handlers: handlers}
}

// RUN PIPELINE HANDLER TESTS

func TestHandlers_RunPipeline_HappyPath(t *testing.T) {
	suite := NewSuite()

	pipeline := models.RunPipelineRequest{
		RepositoryUrl: "ysayonnar/pipecraft",
		Branch:        "main",
		Commit:        "e4r3e2",
	}

	requestBody, _ := json.Marshal(pipeline)
	req, _ := http.NewRequest(http.MethodPost, "/run-pipeline", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	suite.handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var response models.RunPipelineResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, response.PipelineId, int64(1))

	// checking if handler detects similar pipeline
	req, _ = http.NewRequest(http.MethodPost, "/run-pipeline", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	suite.handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, response.PipelineId, int64(1))
}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestHandlers_RunPipeline_MethodNotAllowed_EmptyBody(t *testing.T) {
	suite := NewSuite()

	// testing method not allowed
	req, _ := http.NewRequest(http.MethodGet, "/run-pipeline", nil)
	rr := httptest.NewRecorder()
	suite.handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// testing reading json
	req, _ = http.NewRequest(http.MethodPost, "/run-pipeline", errorReader{})
	rr = httptest.NewRecorder()
	suite.handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)

	// testing parsing json
	req, _ = http.NewRequest(http.MethodPost, "/run-pipeline", bytes.NewBuffer([]byte("")))
	rr = httptest.NewRecorder()
	suite.handlers.RunPipeline(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlers_RunPipeline_ErrorPipelineService(t *testing.T) {
	redisMock := NewMockRedisServie()
	pipelinesMock := NewErrorMockPipelineService()
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
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
