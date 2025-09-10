package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"pipecraft/internal/models"
	"pipecraft/internal/storage"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func NewSuiteWithPipeline() (*Suite, int64) {
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

	data, err := io.ReadAll(rr.Body)
	if err != nil {
		panic(err)
	}

	var responseJson models.RunPipelineResponse
	err = json.Unmarshal(data, &responseJson)
	if err != nil {
		panic(err)
	}

	return suite, responseJson.PipelineId
}

func TestHandlers_PipelineStatus_HappyPath(t *testing.T) {
	suite, pipelineId := NewSuiteWithPipeline()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	data, err := io.ReadAll(rr.Body)
	require.NoError(t, err)

	var responseJson models.PipelineStatusResponse
	err = json.Unmarshal(data, &responseJson)
	require.NoError(t, err)

	require.Equal(t, storage.PIPELINE_STATUS_WAITING, responseJson.Status)
}

func TestHandlers_PipelineStatus_MethodNotAllowed_EmptyParams(t *testing.T) {
	suite, pipelineId := NewSuiteWithPipeline()

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr = httptest.NewRecorder()

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req, _ = http.NewRequest(http.MethodGet, "/pipeline/smth/status", nil)
	rr = httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": "smth"})

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlers_PipelineStatus_CachedResponse(t *testing.T) {
	suite, pipelineId := NewSuiteWithPipeline()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// cached response
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr = httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestHandlers_PipelineStatus_PipelineService_Error(t *testing.T) {
	suite := NewSuite()

	pipelineId := 1

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)

	redisService := NewMockRedisServie()
	errorPipelineService := NewErrorMockPipelineService()
	handlers := New(redisService, errorPipelineService)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr = httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	handlers.PipelineStatus(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
