package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestHandlers_PipelineLogs_HappyPath(t *testing.T) {
	suite, pipelineId := NewSuiteWithPipeline()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/logs", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestHandlers_PipelineLogs_MethodNotAllowed_EmptyParams(t *testing.T) {
	suite, pipelineId := NewSuiteWithPipeline()

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/pipeline/%d/logs", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/logs", pipelineId), nil)
	rr = httptest.NewRecorder()

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/logs", pipelineId), nil)
	rr = httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": "smth"})

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestHandlers_PipelineLogs_CachedResponse(t *testing.T) {
	suite, pipelineId := NewSuiteWithPipeline()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/logs", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// cached response
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/logs", pipelineId), nil)
	rr = httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestHandlers_PipelineLogs_PipelineServiceError(t *testing.T) {
	redisService := NewMockRedisServie()
	errorPipelineService := NewErrorMockPipelineService()
	handlers := New(redisService, errorPipelineService)

	pipelineId := 1

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandlers_PipelineLogs_NotFound(t *testing.T) {
	suite := NewSuite()

	pipelineId := 1

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/pipeline/%d/status", pipelineId), nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(int(pipelineId))})

	suite.handlers.PipelineLogs(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}
